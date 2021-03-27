package storage

import (
	"context"
)

type Script struct {
	Uuid      string `gorm:"primaryKey"`
	Created   string
	Updated   string
	Author    string
	Signature string
	Body      string
}

type Server struct {
	Uuid   string `gorm:"primaryKey"`
	Host   string
	Port   int
	Status string
	Weight int
}

type Property struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

func initCommonTables(ctx context.Context) error {
	if db := db(ctx); db != nil {
		return db.AutoMigrate(
			&Script{},
			&Server{},
			&Property{},
		)
	}
	panic("WHY")
	return nil
}
