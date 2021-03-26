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
