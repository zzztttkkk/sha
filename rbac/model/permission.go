package model

import (
	x "github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/sqlx"
)

type Permission struct {
	sqlx.Model
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (Permission) TableName() string {
	return TablenamePrefix + "permission"
}

func (perm Permission) TableColumns(db *x.DB) []string {
	return append(
		perm.Model.TableColumns(db),
		"name varchar(512) unique",
		"description text",
	)
}
