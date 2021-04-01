package cluster

import (
	"context"
	"encoding/json"
	"time"

	"github.com/apex/log"
	"github.com/hashicorp/memberlist"
	"github.com/spf13/viper"

	"jabberwocky/storage"
)

var (
	mlist      *memberlist.Memberlist
	dconf      *memberlist.Config
	broadcasts *memberlist.TransmitLimitedQueue
	handler    delegate
)

type MemberEvent struct {
	Server storage.Server
	Leave  bool
}

func startGossip(ctx context.Context, eventChan chan MemberEvent) error {
	nodeId, err := storage.GetNodeId(ctx)
	if err != nil {
		return err
	}

	handler.ctx = ctx
	handler.nodeId = nodeId
	handler.memberEvents = eventChan

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
}

type delegate struct {
	nodeId       string
	ctx          context.Context
	memberEvents chan MemberEvent
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

	state.Status = nodeState(node)

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

	d.memberEvents <- MemberEvent{state, true}
}

func (d *delegate) NotifyUpdate(node *memberlist.Node) {
	log.Info("A node was updated: " + node.String())
}

type broadcast struct {
	msg []byte
}

func castMsg(msg []byte) *broadcast {
	return &broadcast{msg: msg}
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
