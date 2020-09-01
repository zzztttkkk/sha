package rbac

import (
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/jsonx"
	"github.com/zzztttkkk/suna/sqls"
)

type logT struct {
	sqls.Model
	Name     string       `json:"Name"`
	Operator int64        `json:"operator"`
	Info     jsonx.Object `json:"info"`
}

func (logT) TableName() string {
	return tablePrefix + "log"
}

func (log logT) TableDefinition(db *sqlx.DB) []string {
	return log.Model.TableDefinition(
		db,
		"Name char(30) not null",
		"operator bigint not null",
		"info json",
	)
}
