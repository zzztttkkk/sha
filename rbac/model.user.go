package rbac

import (
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/sqls"
)

type userWithRoleT struct {
	sqls.Model
	Subject int64 `json:"subject"`
	Role    int64 `json:"role"`
}

func (userWithRoleT) TableName() string {
	return tablePrefix + "subject_with_role"
}

func (ele userWithRoleT) SqlsTableColumns(db *sqlx.DB) []string {
	return ele.Model.SqlsTableColumns(
		db,
		"subject bigint not null",
		"role bigint not null",
	)
}
