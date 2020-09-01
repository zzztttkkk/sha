package rbac

import (
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/sqls"
)

type permissionT struct {
	sqls.Enum
	Descp string ` json:"descp"`
}

func (permissionT) TableName() string {
	return tablePrefix + "permission"
}

func (perm permissionT) TableDefinition(db *sqlx.DB) []string {
	return perm.Enum.TableDefinition(db, "descp text")
}
