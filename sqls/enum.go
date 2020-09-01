package sqls

import "github.com/jmoiron/sqlx"

type EnumItem interface {
	GetId() int64
	GetName() string
}

type Enum struct {
	Model
	Name string `json:"name"`
}

func (enum *Enum) GetId() int64 {
	return enum.Id
}

func (enum *Enum) GetName() string {
	return enum.Name
}

func (enum Enum) TableDefinition(db *sqlx.DB, lines ...string) []string {
	return enum.Model.TableDefinition(db, append(lines, "name char(255) not null unique")...)
}
