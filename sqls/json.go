package sqls

import (
	"github.com/zzztttkkk/suna/sqls/internal"
)

var JsonSet func(column string, path string, v interface{}) internal.Sqlizer

var JsonUpdate func(column string, m map[string]interface{}) internal.Sqlizer

var JsonRemove func(column string, paths ...string) internal.Sqlizer
