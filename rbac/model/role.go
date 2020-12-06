package model

import (
	x "github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/sqlx"
)

type Role struct {
	sqlx.Model
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Permissions []*Permission `json:"permissions" db:"-"`
}

func (Role) TableName() string {
	return TablenamePrefix + "role"
}

func (r Role) TableColumns(db *x.DB) []string {
	return append(
		r.Model.TableColumns(db),
		"name varchar(512) unique",
		"description text",
	)
}

type RoleWithPermissions struct {
	Role int64
	Perm int64
}

func (rwp RoleWithPermissions) TableName() string {
	return TablenamePrefix + "role_with_perms"
}

func (rwp RoleWithPermissions) TableColumns(_ *x.DB) []string {
	return []string{
		"role bigint not null",
		"perm bigint not null",
		"primary key(role,perm)",
	}
}

type RoleWithInheritances struct {
	Role  int64
	Based int64 `db:"based"`
}

func (rwp RoleWithInheritances) TableName() string {
	return TablenamePrefix + "role_with_inheritances"
}

func (rwp RoleWithInheritances) TableColumns(_ *x.DB) []string {
	return []string{
		"role bigint not null",
		"based bigint not null",
		"primary key(role,based)",
	}
}
