package storage

import (
	"context"
)

type Script struct {
	Uuid      string
	Created   string
	Updated   string
	Author    string
	Signature string
	Body      string
}

type Server struct {
	Uuid   string
	Host   string
	Port   int
	Status string
	Weight string
}

func initCommonTables(ctx context.Context) error {
	if db := db(ctx); db != nil {
		return db.AutoMigrate(
			&Script{},
			&Server{},
		)
	}
	panic("WHY")
	return nil
}
