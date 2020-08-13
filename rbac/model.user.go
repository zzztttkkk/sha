package rbac

import "github.com/zzztttkkk/suna/sqls"

type userWithRoleT struct {
	sqls.Model
	Subject int64 `json:"subject"`
	Role    int64 `json:"role"`
}

func (userWithRoleT) TableName() string {
	return tablePrefix + "subject_with_role"
}

func (ele userWithRoleT) TableDefinition() []string {
	return ele.Model.TableDefinition(
		"subject bigint not null",
		"role bigint not null",
	)
}
