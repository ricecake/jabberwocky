package cluster

import (
	"context"
	"encoding/json"

	"jabberwocky/storage"

	"github.com/apex/log"
)

// TODO: make this accept a channel for "cluster events" so that we can push node up/down to clients when they happen
func StartCluster(ctx context.Context, eventChan chan MemberEvent) error {
	if err := startGossip(ctx, eventChan); err != nil {
		return err
	}

	if err := startDiscovery(ctx); err != nil {
		return err
	}

	return nil
}

func Servers() (nodes []storage.Server) {
	for _, mem := range mlist.Members() {
		var state storage.Server
		if err := json.Unmarshal(mem.Meta, &state); err != nil {
			log.Error(err.Error())
			return
		}

		state.Status = nodeState(mem)
		nodes = append(nodes, state)
	}
	return
}

func Health() string {
	return nodeState(mlist.LocalNode())
}

func Priority() int {
	return mlist.GetHealthScore()
}
