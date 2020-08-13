package rbac

import "github.com/zzztttkkk/suna/sqls"

type permissionT struct {
	sqls.Enum
	Descp string ` json:"descp"`
}

func (permissionT) TableName() string {
	return tablePrefix + "permission"
}

func (perm permissionT) TableDefinition() []string {
	return perm.Enum.TableDefinition("descp text")
}
