package storage

import "context"

type Key struct {
	PublicKey  string
	PrivateKey string
}

func initAgentTables(ctx context.Context) error {
	if db := db(ctx); db != nil {
		db.AutoMigrate(
			&Key{},
		)
	}

	return nil
}
