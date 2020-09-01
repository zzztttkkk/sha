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

func (perm permissionT) SqlsTableColumns(db *sqlx.DB) []string {
	return perm.Enum.SqlsTableColumns(db, "descp text")
}
