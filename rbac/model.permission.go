package rbac

import "github.com/zzztttkkk/suna/sqls"

type _Permission struct {
	sqls.Enum
	Descp string ` json:"descp"`
}

func (_Permission) TableName() string {
	return tablePrefix + "permission"
}

func (perm _Permission) TableDefinition() []string {
	return perm.Enum.TableDefinition("descp text")
}

type _PermConflict struct {
	A int64
	B int64
}

func (_PermConflict) TableName() string {
	return tablePrefix + "perm_conflict"
}

func (_PermConflict) TableDefinition() []string {
	return []string{
		"a bigint not null",
		"b bigint not null",
		"primary key(a,b)",
	}
}
