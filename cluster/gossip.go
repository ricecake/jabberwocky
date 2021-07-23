package cluster

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/hashicorp/memberlist"
	"github.com/spf13/viper"

	"jabberwocky/storage"
	"jabberwocky/transport"
)

var (
	mlist      *memberlist.Memberlist
	dconf      *memberlist.Config
	broadcasts *memberlist.TransmitLimitedQueue
	handler    delegate
)

// This should turn into a generic "new data" message, so "Events", and can be used to relay events as they happen, with sync being used to ensure consistency.
// The big focus should be on making sure the clients know when a new agent connects.  But all composition events, including jobs and whatnot, need to get forwarded to the cluster, and to the clients.  Should consider something like "AnnounceAgent/Client"? Maybe something with a transport message?

// NEED TO MAKE ALL CLUSTER MESSAGES BE SIGNED clusterFrame jwts

func startGossip(ctx context.Context) error {
	nodeId, err := storage.GetNodeId(ctx)
	if err != nil {
		return err
	}

	handler.ctx = ctx
	handler.nodeId = nodeId
	handler.nodeMap = make(map[string]map[string]bool)
	handler.lock = &sync.RWMutex{}

	dconf = memberlist.DefaultLANConfig()
	dconf.BindPort = viper.GetInt("server.gossip_port")
	dconf.Name = nodeId
	dconf.Events = &handler
	dconf.Delegate = &handler

	list, err := memberlist.Create(dconf)
	if err != nil {
		return err
	}
	mlist = list

	broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return mlist.NumMembers()
		},
		RetransmitMult: 3,
	}

	go func() {
		log.Info("Starting broadcast loop")
		for msg := range Router.GetClusterOutbound() {
			log.Infof("broadcast %+v", msg)
			cast, err := castMsg(msg, client)
			if err != nil {
				log.Error(err.Error())
				continue
			}

			broadcasts.QueueBroadcast(cast)
		}
		log.Info("Leaving broadcast loop")
	}()

	go func() {
		select {
		case <-ctx.Done():
			shutdownGossip()
		}
	}()

	log.Info("Started gossip")

	return nil
}

func shutdownGossip() {
	mlist.Leave(time.Duration(5) * time.Second)
	mlist.Shutdown()
	log.Info("Gossip stopped")
}

type NodeState struct {
	State   storage.Server
	Agents  []storage.Agent  `json:"omitempty"`
	Scripts []storage.Script `json:"omitEmpty"`
	// This should also include a map of subscriptions, both for this node, and for all known node subscriptions.
	// it should also include any persistent jobs, and known agents and their details.
}

type delegate struct {
	nodeId  string
	ctx     context.Context
	lock    *sync.RWMutex
	nodeMap map[string]map[string]bool
}

func (d *delegate) NodeMeta(limit int) []byte {
	server, err := storage.GetServer(d.ctx, d.nodeId)
	if err != nil {
		log.Error(err.Error())
		return []byte{}
	}

	data, err := json.Marshal(&server)
	if err != nil {
		log.Error(err.Error())
	}

	return data
}

func (d *delegate) NotifyMsg(b []byte) {
	var frame clusterEnvelope
	err := json.Unmarshal(b, &frame)
	if err != nil {
		log.Error(err.Error())
		return
	}

	d.lock.RLock()
	msgMap, found := d.nodeMap[frame.Server]
	d.lock.RUnlock()
	if !found {
		msgMap = make(map[string]bool)
		d.lock.Lock()
		d.nodeMap[frame.Server] = msgMap
		d.lock.Unlock()
	}

	d.lock.RLock()
	seen := msgMap[frame.Id]
	d.lock.RUnlock()
	if !seen {
		d.lock.Lock()
		msgMap[frame.Id] = true
		d.lock.Unlock()

		Router.Emit(frame.Emitter.ConvertToPeer(), frame.Message)
	}
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return broadcasts.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	var state NodeState

	server, err := storage.GetServer(d.ctx, d.nodeId)
	if err != nil {
		log.Error(err.Error())
		return []byte{}
	}

	state.State = server

	data, err := json.Marshal(&state)
	if err != nil {
		log.Error(err.Error())
	}

	return data
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	var state NodeState
	if err := json.Unmarshal(buf, &state); err != nil {
		log.Error(err.Error())
		return
	}

	d.lock.Lock()
	d.nodeMap[state.State.Uuid] = make(map[string]bool)
	d.lock.Unlock()

	if err := storage.SaveServer(d.ctx, state.State); err != nil {
		log.Error(err.Error())
		return
	}
}

func (d *delegate) NotifyJoin(node *memberlist.Node) {
	log.Info("A node has joined: " + node.String())

	var state storage.Server
	if err := json.Unmarshal(node.Meta, &state); err != nil {
		log.Error(err.Error())
		return
	}

	if state.Uuid != d.nodeId {
		channel := Router.RegisterPeer(state.Uuid)
		go func() {
			for msg := range channel {
				cast, err := castMsg(msg, agent)
				if err != nil {
					log.Error(err.Error())
					continue
				}
				mlist.SendReliable(node, cast.Message())
			}
		}()
	}

	d.lock.Lock()
	d.nodeMap[state.Uuid] = make(map[string]bool)
	d.lock.Unlock()

	state.Status = nodeState(node)

	// this should probably be sending an event to the cluster.Router object.
	Router.Emit(PEER_SERVER, transport.NewMessage("server", "join", state))
	// d.memberEvents <- MemberEvent{state, false}
}

func (d *delegate) NotifyLeave(node *memberlist.Node) {
	log.Info("A node has left: " + node.String())

	var state storage.Server
	if err := json.Unmarshal(node.Meta, &state); err != nil {
		log.Error(err.Error())
		return
	}

	state.Status = "offline"
	d.lock.Lock()
	delete(d.nodeMap, state.Uuid)
	d.lock.Unlock()
	Router.UnregisterPeer(state.Uuid)

	// this should probably be sending an event to the cluster.Router object.
	Router.Emit(PEER_SERVER, transport.NewMessage("server", "leave", state))
	// d.memberEvents <- MemberEvent{state, true}
}

func (d *delegate) NotifyUpdate(node *memberlist.Node) {
	log.Info("A node was updated: " + node.String())
}

type frameType int

const (
	client frameType = iota
	agent
)

type clusterFrame struct { // This should get updated to be basically a jwt, so it can be signed.
	Id      string
	Node    string
	Type    frameType
	Message transport.Message
}

type broadcast struct {
	msg []byte
}

func castMsg(msg clusterEnvelope, t frameType) (*broadcast, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return &broadcast{
		msg: data,
	}, nil
}

func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {}

func nodeState(peer *memberlist.Node) string {
	switch peer.State {
	case memberlist.StateAlive:
		return "alive"
	default:
		return "degraded"
	}
}
