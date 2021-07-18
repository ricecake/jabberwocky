package cluster

import (
	"context"
	"encoding/json"
	"time"

	"github.com/apex/log"
	"github.com/hashicorp/memberlist"
	"github.com/ricecake/karma_chameleon/util"
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
type MemberEvent struct {
	Server storage.Server
	Leave  bool
}

// NEED TO MAKE ALL CLUSTER MESSAGES BE SIGNED clusterFrame jwts

func startGossip(ctx context.Context, eventChan chan MemberEvent) error {
	nodeId, err := storage.GetNodeId(ctx)
	if err != nil {
		return err
	}

	handler.ctx = ctx
	handler.nodeId = nodeId
	handler.memberEvents = eventChan
	handler.msgMap = make(map[string]map[string]bool)

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
			close(eventChan)
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
	nodeId       string
	ctx          context.Context
	memberEvents chan MemberEvent
	msgMap       map[string]map[string]bool
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
	var frame clusterFrame
	err := json.Unmarshal(b, &frame)
	if err != nil {
		log.Error(err.Error())
		return
	}
	ctxLog := log.WithFields(log.Fields{
		"node": frame.Node,
		"id":   frame.Id,
	})

	ctxLog.Info("Broadcast")

	nodeMap, found := d.msgMap[frame.Node]
	if !found {
		ctxLog.Info("Initializing")
		nodeMap = make(map[string]bool)
		d.msgMap[frame.Node] = nodeMap
	}
	if !nodeMap[frame.Id] {
		ctxLog.Info("Forwarding")
		nodeMap[frame.Id] = true
		switch frame.Type {
		case client:
			Router.HandleClusterClientInbound(frame.Message)
		case agent:
			Router.HandleClusterAgentInbound(frame.Message)
		default:
			ctxLog.Info("What is this?")
		}
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

	d.msgMap[state.State.Uuid] = make(map[string]bool)

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
				log.Infof("SENDING TO PEER [[%+v]]", msg)
				mlist.SendReliable(node, cast.Message())
			}
		}()
	}

	d.msgMap[state.Uuid] = make(map[string]bool)

	state.Status = nodeState(node)

	// this should probably be sending an event to the cluster.Router object.
	d.memberEvents <- MemberEvent{state, false}
}

func (d *delegate) NotifyLeave(node *memberlist.Node) {
	log.Info("A node has left: " + node.String())

	var state storage.Server
	if err := json.Unmarshal(node.Meta, &state); err != nil {
		log.Error(err.Error())
		return
	}

	state.Status = "offline"
	delete(d.msgMap, state.Uuid)
	Router.UnregisterPeer(state.Uuid)

	// this should probably be sending an event to the cluster.Router object.
	d.memberEvents <- MemberEvent{state, true}
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

func castMsg(msg transport.Message, t frameType) (*broadcast, error) {
	data, err := json.Marshal(&clusterFrame{
		Id:      util.CompactUUID(),
		Type:    t,
		Node:    handler.nodeId,
		Message: msg,
	})
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
