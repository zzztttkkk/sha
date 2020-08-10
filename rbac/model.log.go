package rbac

import (
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
)

type Log struct {
	sqls.Model
	Name     string           `json:"name"`
	Operator int64            `json:"operator"`
	Info     utils.JsonObject `json:"info"`
}

func (Log) TableName() string {
	return tablePrefix + "log"
}

func (log Log) TableDefinition() []string {
	return log.Model.TableDefinition(
		"name char(30) not null",
		"operator bigint not null",
		"info json",
	)
}
