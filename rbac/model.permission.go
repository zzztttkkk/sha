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
