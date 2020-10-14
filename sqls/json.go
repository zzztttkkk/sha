package sqls

import (
	"github.com/zzztttkkk/suna/sqls/internal"
)

var _JsonSetImpl func(column string, path string, v interface{}) internal.Sqlizer

var _JsonUpdateImpl func(column string, m map[string]interface{}) internal.Sqlizer

var _JsonRemoveImpl func(column string, paths ...string) internal.Sqlizer

func JsonSet(column string, path string, v interface{}) internal.Sqlizer {
	return _JsonSetImpl(column, path, v)
}

func JsonUpdate(column string, m map[string]interface{}) internal.Sqlizer {
	return _JsonUpdateImpl(column, m)
}

func JsonRemove(column string, paths ...string) internal.Sqlizer {
	return _JsonRemoveImpl(column, paths...)
}
