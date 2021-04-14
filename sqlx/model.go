package sqlx

import "github.com/jmoiron/sqlx"

type Model struct {
	ID      int64 `json:"id" db:"id,immutable,g=info"`
	Status  int   `json:"status" db:"status"`
	Created int64 `json:"created" db:"created,immutable,g=info"`
	Deleted int64 `json:"-" db:"deleted"`
}

func (Model) TableColumns(db *sqlx.DB) []string {
	var s []string
	idLine := ""
	switch db.DriverName() {
	case "postgres", "pgx":
		idLine = "`id` serial8 primary key"
	case "sqlite3":
		idLine = "`id` integer primary key"
	case "mysql":
		idLine = "`id` bigint not null auto_increment primary key"
	}

	s = append(s, idLine)
	s = append(
		s,
		"`status` int default 0",
		"`created` bigint unsigned not null",
		"`deleted` bigint unsigned default 0",
	)
	return s
}

type Modeler interface {
	TableName() string
	TableColumns(db *sqlx.DB) []string
}
