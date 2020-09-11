package rbac

import (
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/sqls"
)

type Role struct {
	sqls.Enum

	Based       []int64 `json:"based" db:"-"`
	Permissions []int64 `json:"permissions" db:"-"`
}

func (Role) SqlsTableName() string {
	return tablePrefix + "role"
}

func (role Role) SqlsTableColumns(db *sqlx.DB) []string {
	return role.Enum.SqlsTableColumns(db)
}

type roleInheritanceT struct {
	Role  int64 `json:"role"`
	Based int64 `json:"based"`
}

func (roleInheritanceT) SqlsTableName() string {
	return tablePrefix + "role_inheritance"
}

func (ele roleInheritanceT) SqlsTableColumns() []string {
	return []string{
		"role bigint not null",
		"based bigint not null",
		"primary key(role, based)",
	}
}

type roleWithPermT struct {
	Role int64 `json:"role"`
	Perm int64 `json:"perm"`
}

func (roleWithPermT) SqlsTableName() string {
	return tablePrefix + "role_with_perm"
}

func (ele roleWithPermT) SqlsTableColumns() []string {
	return []string{
		"role bigint not null",
		"perm bigint not null",
		"primary key(role, perm)",
	}
}
