package cluster

import (
	"context"

	"github.com/apex/log"
	"github.com/hashicorp/memberlist"
	"github.com/pborman/uuid"
	"github.com/spf13/viper"
)

var (
	mlist *memberlist.Memberlist
	dconf *memberlist.Config
)

func startGossip(ctx context.Context) error {
	dconf = memberlist.DefaultLANConfig()
	dconf.BindPort = viper.GetInt("server.gossip_port")
	dconf.Name = uuid.NewRandom().String()

	list, err := memberlist.Create(dconf)
	if err != nil {
		return err
	}
	mlist = list

	log.Info("Started gossip")

	return nil
}
