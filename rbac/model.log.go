package rbac

import (
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

func (log logT) TableDefinition() []string {
	return log.Model.TableDefinition(
		"Name char(30) not null",
		"operator bigint not null",
		"info json",
	)
}
