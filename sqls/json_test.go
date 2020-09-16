package sqls

import (
	"fmt"
	"testing"

	"github.com/zzztttkkk/suna/jsonx"
)

func TestMysqlJson(t *testing.T) {
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
