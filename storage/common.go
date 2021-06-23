package storage

import (
	"context"
	"fmt"
	"net/url"
)

type Script struct {
	Uuid      string `gorm:"primaryKey"`
	Created   string
	Updated   string
	Author    string
	Signature string
	Body      string
	Security  int // 0: internal state access only, no host access.  1: "preformated" host access, status etc.  2: read only host access. 3: r/w host access 4: command execution
}

type Server struct {
	Uuid   string `gorm:"primaryKey"`
	Host   string
	Port   int
	Status string
	Weight int
}

func (srv Server) Url() url.URL {
	host := srv.Host
	if srv.Port != 0 && srv.Port != 443 {
		host = fmt.Sprintf("%s:%d", host, srv.Port)
	}
	return url.URL{
		Scheme: "wss",
		Host:   host,
	}
}

func (srv Server) HashKey() string {
	return srv.Uuid
}

func (srv Server) HashWeight() int {
	return srv.Weight + 1
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
