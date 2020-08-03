package rbac

import "github.com/zzztttkkk/suna/sqls"

type _UserWithRole struct {
	sqls.Model
	Subject  int64 `json:"subject"`
	Role     int64 `json:"role"`
	Operator int64 `json:"operator"`
}

func (_UserWithRole) TableName() string {
	return tablePrefix + "subject_with_role"
}

func (ele _UserWithRole) TableDefinition() []string {
	return ele.Model.TableDefinition(
		"subject bigint not null",
		"role bigint not null",
		"operator bigint not null",
	)
}
