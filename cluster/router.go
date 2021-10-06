package cluster

import (
	"sync"

	"github.com/davecgh/go-spew/spew"

	"jabberwocky/transport"
)

/*
Should also make this handle node join/leave messages, since those are tightly coupled with cluster management.
Attach/detach of clients and agents is better handled at the server level, since we at most care about transmitting state change
at a broadcast level.

Then this can be something thats either passed in, or returned from the startCluster methods.


Jobs and all state, like scripts, should be shared via normal gossip procedures.
Broadcast should still be used when creating a new job, sending a command, and the like, but
persistant things should always be in gossip, since a broadcast isnt certain to succeed.

Also need to create an Envelope type, since the transport messages are suitable for agent-server, but more context
is helpful for server-server, for logging of responsible user, and origin server.

When creating an envelope, it can examine the message it holds, and fill in most of the fields.  Tags are most important.
*/

//THESE SHOULD BE SYNC RWMutex
type router struct {
	peerLock   *sync.RWMutex
	clientLock *sync.RWMutex
	agentLock  *sync.RWMutex

	subRouter *SubsetRouter

	clusterOutbound    chan clusterEnvelope
	processingOutbound chan transport.Message
	storageOutbound    chan transport.Message

	peerOutbound   map[string]chan clusterEnvelope
	clientOutbound map[string]chan transport.Message
	agentOutbound  map[string]chan transport.Message
}

func NewRouter() *router {
	return &router{
		peerLock:   &sync.RWMutex{},
		clientLock: &sync.RWMutex{},
		agentLock:  &sync.RWMutex{},

		subRouter: NewSubsetRouter(),

		clusterOutbound:    make(chan clusterEnvelope, 1),
		processingOutbound: make(chan transport.Message, 1),
		storageOutbound:    make(chan transport.Message, 1),

		peerOutbound:   make(map[string]chan clusterEnvelope),
		clientOutbound: make(map[string]chan transport.Message),
		agentOutbound:  make(map[string]chan transport.Message),
	}
}

func (r *router) Dump() {
	spew.Dump(r)
}

func (r *router) GetClusterOutbound() chan clusterEnvelope {
	return r.clusterOutbound
}

func (r *router) GetStorageOutbound() chan transport.Message {
	return r.storageOutbound
}

func (r *router) GetProcessingOutbound() chan transport.Message {
	return r.processingOutbound
}

func (r *router) RegisterAgent(code string) chan transport.Message {
	r.agentLock.Lock()
	defer r.agentLock.Unlock()

	if agentChan, found := r.agentOutbound[code]; found {
		return agentChan
	}

	agentChan := make(chan transport.Message)
	r.agentOutbound[code] = agentChan
	return agentChan
}

func (r *router) UnregisterAgent(code string) {
	r.agentLock.Lock()
	defer r.agentLock.Unlock()

	if agentChan, found := r.agentOutbound[code]; found {
		close(agentChan)
		delete(r.agentOutbound, code)
	}
}

func (r *router) RegisterClient(code string) chan transport.Message {
	r.clientLock.Lock()
	defer r.clientLock.Unlock()

	if clientChan, found := r.clientOutbound[code]; found {
		return clientChan
	}

	clientChan := make(chan transport.Message)
	r.clientOutbound[code] = clientChan
	return clientChan
}

func (r *router) UnregisterClient(code string) {
	r.clientLock.Lock()
	defer r.clientLock.Unlock()

	if clientChan, found := r.clientOutbound[code]; found {
		close(clientChan)
		delete(r.clientOutbound, code)
	}
}

func (r *router) RegisterPeer(code string) chan clusterEnvelope {
	r.peerLock.Lock()
	defer r.peerLock.Unlock()

	if peerChan, found := r.peerOutbound[code]; found {
		return peerChan
	}

	peerChan := make(chan clusterEnvelope)
	r.peerOutbound[code] = peerChan
	return peerChan
}

func (r *router) UnregisterPeer(code string) {
	r.peerLock.Lock()
	defer r.peerLock.Unlock()

	if peerChan, found := r.peerOutbound[code]; found {
		close(peerChan)
		delete(r.peerOutbound, code)
	}
}

func (r *router) AddAgentBinding(code string, binding map[string]string) {} //Superset binding -- all msg tags in binding

func (r *router) AddClientBinding(code string, binding map[string]string) {
	r.subRouter.AddBind(Destination{
		Role: LOCAL_CLIENT,
		Code: code,
	}, binding)

} // subset binding -- all bindings in msg tags

func (r *router) AddServerBinding(code string, binding map[string]string) {} // subset binding -- all bindings in msg tags

/*
Need a function like this, but for sending to storage.
this is needed because if we gossip a message that needs it's UUID filled in, or other "defaulted" fields,
then we will end up deriving multiple different derived fields on each server.
So when creating a payload, for example, we want to route it to storage, and then let storage announce
that it should be broadcast.
Need to make sure that the handling of messages from the local server/to the peer server get gossiped and persisted correctly.
Probably fine to have the agent/client websocket connections grab anything with the subtype of CRUD verbs and route to storage, else send to emit.
switch msg.SubType{case ... crud: Storage; default: emit}

Local server should broadcast to client, and broadcast to cluster.
peer server should storage, and broadcast to client.
joining/connecting agent/client should move to sending to storage where appropriate.

Need to find a consistent way for ensuring agents are informed about leaving/joining servers, but make things normal otherwise.

Should just put message specific type/subtype handing in the emit function.
Break the bodies of the different cases out into helper functions, so that the deeply nested switch statements don't become confusing.
*/

func (r *router) Send(code string, e Emitter, msg transport.Message) {
	switch e {
	case LOCAL_CLIENT:
		r.clientLock.RLock()
		defer r.clientLock.RUnlock()
		if clientChan, found := r.clientOutbound[code]; found {
			clientChan <- msg
		}
	case LOCAL_AGENT:
		r.agentLock.RLock()
		defer r.agentLock.RUnlock()
		if agentChan, found := r.agentOutbound[code]; found {
			agentChan <- msg
		}
	}
}

// Emit handles all messages.  Might change to "Route"?
// gossip libs need to convert emitter fields from local to peer before passing to Emit
func (r *router) Emit(e Emitter, msg transport.Message) {
	// TODO: need to switch on msg type and subtype, to decide where specifically to route it.  Probably move into a function, with the default being to log an error.
	switch e {
	case LOCAL_CLIENT:
		r.handleLocalClientEmit(e, msg)
	case LOCAL_AGENT:
		r.handleLocalAgentEmit(e, msg)
	case LOCAL_SERVER:
		r.handleLocalServerEmit(e, msg)
	case PEER_CLIENT:
		r.handlePeerClientEmit(e, msg)
	case PEER_AGENT:
		r.handlePeerAgentEmit(e, msg)
	case PEER_SERVER:
		r.handlePeerServerEmit(e, msg)
	}
}

func (r *router) handleLocalClientEmit(e Emitter, msg transport.Message) {
	//send to storage processing
	r.storageOutbound <- msg
	//Route to local agents
	r.routeAgent(e, msg)
}
func (r *router) handleLocalAgentEmit(e Emitter, msg transport.Message) {
	// TODO: make agent messages all, even connection, be local agent.
	// Then can have these always send to storage and processing, and then it should make the broadcast and routing part a bit easier to reason about.

	// send to output handling
	r.processingOutbound <- msg
	// send to storage processing
	r.storageOutbound <- msg
	// Route to cluster
	r.routeCluster(e, msg)
	// route to local clients
	r.routeClient(e, msg)
}
func (r *router) handleLocalServerEmit(e Emitter, msg transport.Message) {
	// Local server is feedback from storage/processing mechanism, and agent/client join leave
	// TODO: this should be made to *only* handle messages produced internally.  So sync messages from storage, mostly.
	// This would move the different join/leave messages into the respective handlers

	if msg.SubType == "sync" {
		// broadcast to local clients
		r.broadcastClient(e, msg)
		//Brodcast to cluster
		r.broadcastCluster(e, msg)
	} else {
		// send to storage processing
		r.storageOutbound <- msg
	}
}

func (r *router) handlePeerClientEmit(e Emitter, msg transport.Message) {
	//send to storage processing
	r.storageOutbound <- msg
	//Route to local agents
	r.routeAgent(e, msg)
}
func (r *router) handlePeerAgentEmit(e Emitter, msg transport.Message) {
	// route to local clients
	r.routeClient(e, msg)
}
func (r *router) handlePeerServerEmit(e Emitter, msg transport.Message) {
	// peer server messages are cluster composition changes
	// send to storage processing
	r.storageOutbound <- msg
	// broadcast to local clients
	r.broadcastClient(e, msg)

	if msg.Type == "server" {
		// broadcast to local agents
		r.broadcastAgent(e, msg)
	}
}

/**
"Route" is used to indicate that it's passed through a filtering mechanism
"Broadcast" indicates that it should go to everything.
Agent and Client messaging is always local
cluster routing requires servers to aggregate the routes of their clients.
agent and client routing is discussed below
**/

func (r *router) broadcastCluster(e Emitter, msg transport.Message) {
	r.clusterOutbound <- packageMessage(e, msg)
}
func (r *router) broadcastClient(e Emitter, msg transport.Message) {
	r.clientLock.RLock()
	defer r.clientLock.RUnlock()
	for _, channel := range r.clientOutbound {
		channel <- msg
	}
}
func (r *router) broadcastAgent(e Emitter, msg transport.Message) {
	r.agentLock.RLock()
	defer r.agentLock.RUnlock()
	for _, channel := range r.agentOutbound {
		channel <- msg
	}
}

func (r *router) routeCluster(e Emitter, msg transport.Message) {
	r.peerLock.RLock()
	defer r.peerLock.RUnlock()
	for _, channel := range r.peerOutbound {
		channel <- packageMessage(e, msg)
	}
}
func (r *router) routeClient(e Emitter, msg transport.Message) {
	r.broadcastClient(e, msg)
}
func (r *router) routeAgent(e Emitter, msg transport.Message) {
	r.broadcastAgent(e, msg)
}

/*
BroadcastCluster
BroadcastAgent
BroadcastClient

GetStorageOutbound()
GetClusterOutbound()
GetPeerOutbound(string)
GetClientOutbound(string) // this implies admin connections will need uuids
GetAgentOutbound(string)

type criteriaNode struct {
	field string
	value string
	recipients []string
	subcriteria []*criteriaNode
}

Need a method on messages that returns the map of "routing tags"
this should only basically work with tag based routing.
that will be the tags, and also the type, subtype, and fields relating to the sub message.

Routing Algorithm:
	set the visit list to be empty
	push the root node onto the list
	while the visit list is not empty:
		pop the first node off of the list
		send the message to all recipients
		for each field in the message
			check if field exists in criteria table
			if it doesnt:
				next field
			if it does:
				check if theres a criteria value that matches
				if there is:
					push that node onto the visit list

theres an opportunity for optimization if we keep track of what fields have been visited for each node on the visit list, so we dont repeat all fields each time.
	this optimization requires sorting the keys and criteria fields, so that things can be consistent and nothing gets missed.
theres a possible simplification where each criteria node is a list instead of a hash, and we just loop through each list, and compare field and value.

For agent selection, we need to keep a map of agent tags, to sets of agents.
Then we grab each set for each tag on the command/message, and do an intersection to get the set to broadcast to.

Need to make sure to deduplicate all the lists, to avoid double broadcast.

Agents use different outbound routing, since instead of subscribing to limit what they get, they anounce criteria to see what theyre eligible for.


Need a function that handles adding a new peer, and saving it in a lookup table of code->channel used for gossip direct routing.
Need to make the function that regesters an agent also register it with properties and whatnot in the lookup table. -- probably seperate function for setting those, rather than all in one.

Need methods to RouteToPeer/Agent/Client.  Messages inbound from agent will route to both peer and client, and inbound from peer will route to client.

The cluster bits should setup a goroutine that holds the actual remote server reference, and also hold the channel that's used to send to that server.
so when a server joins, the peer is registered, which creates the channel, and then sends a message to the member events channel, and returns the channel.  The gossip lib then
starts a go-routine that listens to that channel, and does a reliable/best-effort message to that server of any messages that it gets.











Need to think of a way to more properly handle having methods that say "send this message to all peers.  Send to all relevant agents.  Send to all relevant clients."
Might need to change the methods to be less about the "semantic" meaning of them, like "RouteCluster", and more to the entent from a user perspective.

BroadcastClients -> BroadcastPeers, routeClients
etc, etc.

clusterFrames might need more about the meaning of the message.  "this is to clients. this is to peers".  Maybe "this is informative, this is data, this is a command"?


The register functions should also take objects of what is being registered.  That way it can also emit events about membership changes.
Might be easier though to just make a new function to "IntroduceAgent", "IntroduceServer", and have those do the routing.
Also need to add functions that will dtrt when given a message, based on what type it is, and where it's coming from.
Need to buuld out enums for client/server/agent/peer.  Peer is being used as "other server", while server is "this server".  Not sure the distinction is needed.

The introduction path is good, since it helps consolidate the logic abut where messages get sent into the router, which is where it belongs.
ConnectClient
ConnectAgent
ConnectServer
DisconnectClient
DisconnectAgent
DisconnectServer

HandleInbound(source Emitter, msg transport.Message)

Maybe the actual answer is that when one of those events happens, we just send the message, and include the right type and data to make it make sense?
Emit(source Emitter, msg transport.Message) and then we just has a "join/leave" message type, with subtypes "agent/server"
We always emit the message inside a cluster frame, so that we can include info about where it came from.
a frame from a local agent should be routed to local clients and routed to servers.
a frame from a peer agent should be routed to local clients
a frame from a local client should be broadcast to peer servers, and routed to local agents and consumed locally
a frame from a peer client should be routed to local agents and consumed locally

*/
