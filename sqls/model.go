package sqls

type Model struct {
	Id      int64 `json:"id"`
	Status  int   `json:"status"`
	Created int64 `json:"created"`
	Deleted int64 `json:"deleted"`
}

func (Model) TableDefinition(lines ...string) []string {
	idLine := "id bigint not null auto_increment primary key"
	if isPostgres {
		idLine = "id serial8 primary key"
	}
	return append(
		lines,
		idLine,
		"status int default 0",
		"created bigint not null",
		"deleted bigint default 0",
	)
}
