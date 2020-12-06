package sqlx

import "github.com/jmoiron/sqlx"

type Model struct {
	ID      int64 `json:"id" db:"id"`
	Status  int   `json:"status" db:"status"`
	Created int64 `json:"created" db:"created_at"`
	Deleted int64 `json:"deleted" db:"deleted_at"`
}

func (Model) TableColumns(db *sqlx.DB) []string {
	var s []string
	idLine := ""
	switch db.DriverName() {
	case "postgres", "pgx":
		idLine = "id serial8 primary key"
	case "sqlite3":
		idLine = "id integer primary key"
	case "mysql":
		idLine = "id bigint not null auto_increment primary key"
	}

	s = append(s, idLine)
	s = append(
		s,
		"status int default 0",
		"created_at bigint unsigned not null",
		"deleted_at bigint unsigned default 0",
	)
	return s
}
