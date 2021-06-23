package storage

import (
	"context"
	"time"
)

type Agent struct {
	Uuid            string `gorm:"primaryKey"`
	PublicKey       string
	PublicKeyId     string `gorm:"index"`
	Status          string `gorm:"index"`
	DelegatedServer string `gorm:"index"`
	LastContact     time.Time
}

type ApiCredential struct {
	Username string
	Hash     string
}

func initServerTables(ctx context.Context) error {
	if db := db(ctx); db != nil {
		db.AutoMigrate(
			&Agent{},
			&ApiCredential{},
		)
	}

	return nil
}
