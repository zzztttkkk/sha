package rbac

import "github.com/zzztttkkk/suna/sqls"

type Log struct {
	sqls.Model
	Name     string          `json:"name"`
	Operator int64           `json:"operator"`
	Info     sqls.JsonObject `json:"info"`
	Success  bool            `json:"success"`
}

func (Log) TableName() string {
	return tablePrefix + "log"
}

func (log Log) TableDefinition() []string {
	return log.Model.TableDefinition(
		"name char(30) not null",
		"operator bigint not null",
		"success bool not null",
		"info json",
	)
}
