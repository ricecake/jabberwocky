package storage

import (
	"github.com/apex/log"
	"github.com/pborman/uuid"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"context"
	"net/url"
	"strconv"
	"time"
)

func ConnectDb(ctx context.Context) (context.Context, error) {

	viper.SetDefault("global.db", "jabberwocky.db")
	dbFile := viper.GetString("global.db")
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

func SetAgentStatus(ctx context.Context, uuid, status string) error {
	id, err := GetNodeId(ctx)
	if err != nil {
		return err
	}

	return db(ctx).Model(&Agent{Uuid: uuid}).Updates(&Agent{Status: status, DelegatedServer: id, LastContact: time.Now()}).Error
}

func SaveAgent(ctx context.Context, agent Agent) error {
	return db(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "uuid"}},
		DoUpdates: clause.AssignmentColumns([]string{"public_key", "public_key_id", "delegated_server", "last_contat"}),
	}).Create(&agent).Error
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

func SetCurrentServer(ctx context.Context, serv Server) error {
	return db(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&Property{Key: "server_node", Value: serv.Uuid}).Error
}

func GetCurrentServer(ctx context.Context) (Server, error) {
	var serv Server
	dbh := db(ctx)
	err := dbh.Where("uuid = (?)", dbh.Model(&Property{}).Where(&Property{Key: "server_node"}).Select("value")).Find(&serv).Error
	return serv, err
}

func SaveServer(ctx context.Context, serv Server) error {
	return db(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "uuid"}},
		DoUpdates: clause.AssignmentColumns([]string{"host", "port", "status", "weight"}),
	}).Create(&serv).Error
}

func MarkServersUnknown(ctx context.Context) error {
	return db(ctx).Model(&Server{}).Where("status != ?", "unknown").Update("status", "unknown").Error
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

func GetAgent(ctx context.Context, id string) (Server, error) {
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

func ServerFromString(server string) (Server, error) {
	var serv Server
	parsed, err := url.Parse(server)
	if err != nil {
		return serv, err
	}

	port := 443
	if p := parsed.Port(); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil {
			log.Error(err.Error())
		} else {
			port = v
		}
	}

	serv.Host = parsed.Host
	serv.Port = port
	serv.Uuid = parsed.Fragment
	serv.Status = "alive"

	return serv, nil
}
