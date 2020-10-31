package sqls

import (
	"github.com/zzztttkkk/suna/sqls/internal"
)

var jsonSetImpl func(column string, path string, v interface{}) internal.Sqlizer

var jsonUpdateImpl func(column string, m map[string]interface{}) internal.Sqlizer

var jsonRemoveImpl func(column string, paths ...string) internal.Sqlizer

func JsonSet(column string, path string, v interface{}) internal.Sqlizer {
	return jsonSetImpl(column, path, v)
}

func JsonUpdate(column string, m map[string]interface{}) internal.Sqlizer {
	return jsonUpdateImpl(column, m)
}

func JsonRemove(column string, paths ...string) internal.Sqlizer {
	return jsonRemoveImpl(column, paths...)
}
