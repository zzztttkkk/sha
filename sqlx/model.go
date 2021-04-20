package sqlx

import "github.com/jmoiron/sqlx"

type Modeler interface {
	TableName() string
	TableColumns(db *sqlx.DB) []string
}
