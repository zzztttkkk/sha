package sqls

import (
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/zzztttkkk/suna/jsonx"
)

func TestMysqlJson(t *testing.T) {
	initMysqlJson()
	fmt.Println(JsonSet("info", "$.x", 12))
	fmt.Println(
		JsonUpdate(
			"info",
			map[string]interface{}{
				"$.name": "x\"'x",
				"$.age":  45,
				"$.rela": jsonx.Array{
					1, 2, 3,
					jsonx.Object{"xx": 34},
				},
			},
		),
	)
	fmt.Println(
		JsonRemove("info", "$.name", "$.x"),
	)
}

func TestPostgresJson(t *testing.T) {
	initPostgresJson()
	fmt.Println(JsonSet("info", "x", 12))
	fmt.Println(
		JsonUpdate(
			"info",
			map[string]interface{}{
				"name": "x\"'x",
				"age":  45,
				"rela": jsonx.Array{
					1, 2, 3,
					jsonx.Object{"xx": 34},
				},
			},
		),
	)
	fmt.Println(
		JsonRemove("info", "name", "x"),
	)
}
