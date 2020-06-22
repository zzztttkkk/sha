package sqls

type Model struct {
	Id      int64 `json:"id" ddl:"incr;primary"`
	Status  int   `json:"status" ddl:"D<0>"`
	Created int64 `json:"created;const"`
	Deleted int64 `json:"deleted" ddl:"D<0>"`
}
