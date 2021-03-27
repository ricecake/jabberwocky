package cluster

import (
	"context"
	"time"

	"github.com/apex/log"
	"github.com/hashicorp/memberlist"
	"github.com/pborman/uuid"
	"github.com/spf13/viper"
)

var (
	mlist      *memberlist.Memberlist
	dconf      *memberlist.Config
	broadcasts *memberlist.TransmitLimitedQueue
	handler    delegate
)

func startGossip(ctx context.Context) error {
	dconf = memberlist.DefaultLANConfig()
	dconf.BindPort = viper.GetInt("server.gossip_port")
	dconf.Name = uuid.NewRandom().String()
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

type gossipDelegate struct{}

type delegate struct{}

func (d *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (d *delegate) NotifyMsg(b []byte) {
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return broadcasts.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	return []byte{}
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
}

func (d *delegate) NotifyJoin(node *memberlist.Node) {
	log.Info("A node has joined: " + node.String())
}

func (d *delegate) NotifyLeave(node *memberlist.Node) {
	log.Info("A node has left: " + node.String())
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
