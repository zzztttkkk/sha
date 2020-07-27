package rbac

import "github.com/zzztttkkk/suna/sqls"

type Role struct {
	sqls.Enum
	Descp string `json:"descp"`

	Based       []int64 `json:"based" db:"-"`
	Permissions []int64 `json:"permissions" db:"-"`
}

func (Role) TableName() string {
	return tablePrefix + "role"
}

func (role Role) TableDefinition() []string {
	return role.Enum.TableDefinition(
		"parent bigint default 0",
		"descp text",
	)
}

type _RoleInheritance struct {
	Role  int64 `json:"role"`
	Based int64 `json:"based"`
}

func (_RoleInheritance) TableName() string {
	return tablePrefix + "role_inheritance"
}

func (ele _RoleInheritance) TableDefinition() []string {
	return []string{
		"role bigint not null",
		"based bigint not null",
		"primary key(role, perm)",
	}
}

type _RoleWithPerm struct {
	Role int64 `json:"role"`
	Perm int64 `json:"perm"`
}

func (_RoleWithPerm) TableName() string {
	return tablePrefix + "role_with_perm"
}

func (ele _RoleWithPerm) TableDefinition() []string {
	return []string{
		"role bigint not null",
		"perm bigint not null",
		"primary key(role, perm)",
	}
}
