package rbac

import (
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/sqls"
)

type Permission struct {
	sqls.Enum
	Descp string ` json:"descp"`
}

func (Permission) SqlsTableName() string {
	return tablePrefix + "permission"
}

func (perm Permission) SqlsTableColumns(db *sqlx.DB) []string {
	return perm.Enum.SqlsTableColumns(db, "descp text")
}
