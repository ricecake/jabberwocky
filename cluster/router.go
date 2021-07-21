package cluster

import (
	"github.com/apex/log"

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

type router struct {
	processingOutbound chan transport.Message
	storageOutbound    chan transport.Message
	clusterOutbound    chan transport.Message
	peerOutbound       map[string]chan transport.Message
	clientOutbound     map[string]chan transport.Message
	agentOutbound      map[string]chan transport.Message
}

func NewRouter() *router {
	return &router{
		processingOutbound: make(chan transport.Message),
		storageOutbound:    make(chan transport.Message),
		clusterOutbound:    make(chan transport.Message),
		peerOutbound:       make(map[string]chan transport.Message),
		clientOutbound:     make(map[string]chan transport.Message),
		agentOutbound:      make(map[string]chan transport.Message),
	}
}

func (r *router) RegisterAgent(code string) chan transport.Message {
	if agentChan, found := r.agentOutbound[code]; found {
		return agentChan
	}

	agentChan := make(chan transport.Message)
	r.agentOutbound[code] = agentChan
	return agentChan
}

func (r *router) UnregisterAgent(code string) {
	if agentChan, found := r.agentOutbound[code]; found {
		close(agentChan)
		delete(r.agentOutbound, code)
	}
}

func (r *router) RegisterClient(code string) chan transport.Message {
	if clientChan, found := r.clientOutbound[code]; found {
		return clientChan
	}

	clientChan := make(chan transport.Message)
	r.clientOutbound[code] = clientChan
	return clientChan
}

func (r *router) UnregisterClient(code string) {
	if clientChan, found := r.clientOutbound[code]; found {
		close(clientChan)
		delete(r.clientOutbound, code)
	}
}

func (r *router) RegisterPeer(code string) chan transport.Message {
	if peerChan, found := r.peerOutbound[code]; found {
		return peerChan
	}

	peerChan := make(chan transport.Message)
	r.peerOutbound[code] = peerChan
	return peerChan
}

func (r *router) UnregisterPeer(code string) {
	if peerChan, found := r.peerOutbound[code]; found {
		close(peerChan)
		delete(r.peerOutbound, code)
	}
}

func (r *router) HandlePeerInbound(msg transport.Message) error { return nil }

func (r *router) HandleAgentInbound(msg transport.Message) error {
	log.Info("Agent message")
	r.RouteCluster(msg)
	r.RouteClient(msg)
	return nil
}

func (r *router) HandleClientInbound(msg transport.Message) error {
	log.Infof("Got from client: %+v", msg)
	r.BroadcastCluster(msg)
	r.RouteAgent(msg)
	return nil
}

func (r *router) GetClusterOutbound() chan transport.Message {
	return r.clusterOutbound
}

func (r *router) GetStorageOutbound() chan transport.Message {
	return r.clusterOutbound
}

func (r *router) GetProcessingOutbound() chan transport.Message {
	return r.processingOutbound
}

func (r *router) HandleClusterClientInbound(msg transport.Message) error {
	log.Infof("Got from Cluster client: %+v", msg)
	r.RouteAgent(msg)
	return nil
}

func (r *router) HandleClusterAgentInbound(msg transport.Message) error {
	log.Infof("Got from Cluster agent: %+v", msg)
	r.RouteClient(msg)
	return nil
}

func (r *router) BroadcastCluster(msg transport.Message) {
	r.clusterOutbound <- msg
}

func (r *router) BroadcastAgent(msg transport.Message) {
	for code, channel := range r.agentOutbound {
		log.Infof("Broadcasting to agent [%s]", code)
		channel <- msg
	}
}

func (r *router) BroadcastClient(msg transport.Message) {
	for code, channel := range r.clientOutbound {
		log.Infof("Broadcasting to client [%s]", code)
		channel <- msg
	}
}

func (r *router) RouteAgent(msg transport.Message) {
	r.BroadcastAgent(msg)
}
func (r *router) RouteClient(msg transport.Message) {
	r.BroadcastClient(msg)
}
func (r *router) RouteCluster(msg transport.Message) {
	for code, channel := range r.peerOutbound {
		log.Infof("Broadcasting to peer [%s]", code)
		channel <- msg
	}
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

*/
