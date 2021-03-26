package storage

import (
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"context"
)

func ConnectDb(ctx context.Context) (context.Context, error) {
	dbFile := "jabberwocky.db" // This should pull a dir from the config

	if viper.GetBool("debug") {
		dbFile = "file::memory:?cache=shared"
	}
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		return ctx, err
	}

	return context.WithValue(ctx, "dbConn", db), nil
}

func db(ctx context.Context) *gorm.DB {
	if db, ok := ctx.Value("dbConn").(*gorm.DB); ok {
		return db
	}

	return nil
}

func InitTables(ctx context.Context) error {
	if err := initCommonTables(ctx); err != nil {
		return err
	}

	if err := initAgentTables(ctx); err != nil {
		return err
	}

	if err := initServerTables(ctx); err != nil {
		return err
	}

	return nil
}
