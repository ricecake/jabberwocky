package storage

import (
	"context"
	"time"
)

type Agent struct {
	Uuid            string `gorm:"primaryKey"`
	PublicKey       string
	Status          string `gorm:"index"`
	DelegatedServer string `gorm:"index"`
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
