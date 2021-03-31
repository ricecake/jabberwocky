package cluster

import (
	"context"
)

// TODO: make this accept a channel for "cluster events" so that we can push node up/down to clients when they happen
func StartCluster(ctx context.Context) error {
	if err := startGossip(ctx); err != nil {
		return err
	}

	if err := startDiscovery(ctx); err != nil {
		return err
	}

	return nil
}

func Health() string {
	return nodeState(mlist.LocalNode())
}

func Priority() int {
	return mlist.GetHealthScore()
}
