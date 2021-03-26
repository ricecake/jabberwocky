package storage

import (
	"context"
	"time"
)

type Agent struct {
	Uuid            string
	PublicKey       string
	Status          string
	DelegatedServer string
	LastContact     time.Time
}

func initServerTables(ctx context.Context) error {
	if db := db(ctx); db != nil {
		db.AutoMigrate(
			&Agent{},
		)
	}

	return nil
}
