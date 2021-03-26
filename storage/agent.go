package storage

import "context"

type Key struct {
	PublicKey  string
	PrivateKey string
}

type Property struct {
	Name  string
	Value string
}

func initAgentTables(ctx context.Context) error {
	if db := db(ctx); db != nil {
		db.AutoMigrate(
			&Key{},
			&Property{},
		)
	}

	return nil
}
