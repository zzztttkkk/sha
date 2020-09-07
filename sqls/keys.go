package sqlx

import (
	"github.com/zzztttkkk/suna/internal/reflectx"
	"reflect"
	"strings"
)

// get table filed form model type.
func Columns(rt reflect.Type, exclude []string, extra []string) string {
	m := map[string]bool{}
	for _, key := range reflectx.Keys(rt, "db") {
		m[key] = true
	}
	for _, key := range exclude {
		delete(m, key)
	}
	var keys []string
	for k, _ := range m {
		keys = append(keys, k)
	}
	keys = append(keys, extra...)

	return strings.Join(keys, ",")
}
