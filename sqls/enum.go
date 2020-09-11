package sqls

import "github.com/jmoiron/sqlx"

type EnumItem interface {
	GetId() int64
	GetName() string
}

type Enum struct {
	Model
	Name  string `json:"name"`
	Descp string `json:"description"`
}

func (enum *Enum) GetId() int64 {
	return enum.Id
}

func (enum *Enum) GetName() string {
	return enum.Name
}

func (enum Enum) SqlsTableColumns(db *sqlx.DB, lines ...string) []string {
	return enum.Model.SqlsTableColumns(
		db,
		append(
			lines,
			"name varchar(512) not null unique",
			"descp text",
		)...,
	)
}
