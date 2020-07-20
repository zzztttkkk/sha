package sqls

type Model struct {
	Id      int64 `json:"id"`
	Status  int   `json:"status"`
	Created int64 `json:"created"`
	Deleted int64 `json:"deleted"`
}
