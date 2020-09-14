package rbac

import (
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/jsonx"
	"github.com/zzztttkkk/suna/sqls"
)

type logT struct {
	sqls.Model
	Name     string       `json:"name"`
	Operator int64        `json:"operator"`
	Info     jsonx.Object `json:"info"`
}

func (logT) SqlsTableName() string {
	return tablePrefix + "log"
}

func (log logT) SqlsTableColumns(db *sqlx.DB) []string {
	return log.Model.SqlsTableColumns(
		db,
		"name varchar(512) not null",
		"operator bigint not null",
		"info json",
	)
}
