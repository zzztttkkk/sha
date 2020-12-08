package internal

import (
	"database/sql"
	"net/http"
)

var ErrorStatusByValue = map[error]int{}

func init() {
	ErrorStatusByValue[sql.ErrNoRows] = http.StatusNotFound
}
