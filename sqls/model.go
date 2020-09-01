package sqls

import "github.com/jmoiron/sqlx"

type Model struct {
	Id      int64 `json:"id"`
	Status  int   `json:"status"`
	Created int64 `json:"created"`
	Deleted int64 `json:"deleted"`
}

func (Model) TableDefinition(db *sqlx.DB, lines ...string) []string {
	var s []string
	idLine := ""
	switch db.DriverName() {
	case "postgres":
		idLine = "id serial8 primary key"
	case "sqlite3":
		idLine = "id INTEGER PRIMARY KEY"
	case "mysql":
		idLine = "id bigint not null auto_increment primary key"
	}

	s = append(s, idLine)
	s = append(
		s,
		"status int default 0",
		"created bigint not null",
		"deleted bigint default 0",
	)
	return append(s, lines...)
}
