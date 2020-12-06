package model

import (
	x "github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/sqlx"
)

type Log struct {
	sqlx.Model
	Name     string          `json:"name"`
	Operator int64           `json:"operator"`
	Info     sqlx.JsonObject `json:"info"`
}

func (Log) TableName() string {
	return TablenamePrefix + "log"
}

func (log Log) TableColumns(db *x.DB) []string {
	return append(
		log.Model.TableColumns(db),
		"name varchar(512) not null",
		"operator bigint not null",
		"info json",
	)
}
