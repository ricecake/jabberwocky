package cluster

import (
	"context"
)

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
