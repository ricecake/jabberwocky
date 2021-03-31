package storage

import (
	"github.com/pborman/uuid"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"context"
)

func ConnectDb(ctx context.Context) (context.Context, error) {
	dbFile := "jabberwocky.db" // This should pull a dir from the config
	dbLogger := logger.Default

	if viper.GetBool("debug") {
		dbFile = "file::memory:?cache=shared"
		dbLogger.LogMode(logger.Info)
	}
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger: dbLogger,
	})
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

func GetNodeId(ctx context.Context) (string, error) {
	newUUID := uuid.NewRandom().String()
	var prop Property
	err := db(ctx).Where(Property{Key: "node_id"}).Attrs(Property{Value: newUUID}).FirstOrCreate(&prop).Error

	if err != nil {
		return "", err
	}

	return prop.Value, nil
}

func SaveServer(ctx context.Context, serv Server) error {
	return db(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "uuid"}},
		DoUpdates: clause.AssignmentColumns([]string{"host", "port", "status", "weight"}),
	}).Create(&serv).Error
}

func SaveServers(ctx context.Context, servs []Server) error {
	return db(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "uuid"}},
		DoUpdates: clause.AssignmentColumns([]string{"host", "port", "status", "weight"}),
	}).Create(&servs).Error
}

func GetServer(ctx context.Context, id string) (Server, error) {
	var serv Server
	err := db(ctx).Where(Server{Uuid: id}).First(&serv).Error
	return serv, err
}

func ListAllServers(ctx context.Context) ([]Server, error) {
	var servs []Server
	err := db(ctx).Find(&servs).Error
	return servs, err
}

func ListLiveServers(ctx context.Context) ([]Server, error) {
	var servs []Server
	err := db(ctx).Where(Server{Status: "alive"}).Find(&servs).Error
	return servs, err
}
