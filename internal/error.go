package internal

import (
	"database/sql"
	"net/http"
)

var ErrorStatusByValue = map[interface{}]int{}

func init() {
	ErrorStatusByValue[sql.ErrNoRows] = http.StatusNotFound
}
