package sqls

import (
	"fmt"
	"strings"

	"github.com/savsgio/gotils"
	"github.com/zzztttkkk/suna/jsonx"
	"github.com/zzztttkkk/suna/sqls/internal"
)

var (
	mysqlQs  = "'"
	mysqlEqs = "\\'"
)

func mysqlEscape(s string) string { return strings.ReplaceAll(s, mysqlQs, mysqlEqs) }

func marshal(v interface{}) string {
	switch rv := v.(type) {
	case jsonx.Array:
		buf := strings.Builder{}
		buf.WriteString("json_array(")
		end := len(rv) - 1
		for i, v := range rv {
			buf.WriteString(marshal(v))
			if i < end {
				buf.WriteByte(',')
			}
		}
		buf.WriteByte(')')
		return buf.String()
	case jsonx.Object:
		buf := strings.Builder{}
		buf.WriteString("json_object(")

		end := len(rv)
		i := 0
		for k, v := range rv {
			buf.WriteByte('"')
			buf.WriteString(mysqlEscape(k))
			buf.WriteByte('"')
			buf.WriteByte(',')
			buf.WriteString(marshal(v))
			i++
			if i < end {
				buf.WriteByte(',')
			}
		}
		buf.WriteByte(')')
		return buf.String()
	default:
		return gotils.B2S(jsonx.MustMarshal(rv))
	}
}

func mysqlJsonSet(column string, path string, v interface{}) internal.Sqlizer {
	return RAW(
		fmt.Sprintf(
			"json_set(%s, '%s', %s)",
			column,
			mysqlEscape(path),
			marshal(v),
		),
	)
}

func mysqlJsonUpdate(column string, m map[string]interface{}) internal.Sqlizer {
	buf := strings.Builder{}
	buf.WriteString("json_set(")
	buf.WriteString(column)
	buf.WriteByte(',')

	l := len(m)
	i := 0
	for k, v := range m {
		buf.WriteByte('\'')
		buf.WriteString(mysqlEscape(k))
		buf.WriteByte('\'')
		buf.WriteByte(',')

		buf.WriteString(marshal(v))
		i++
		if i < l {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte(')')
	return RAW(buf.String())
}

func mysqlJsonRemove(column string, paths ...string) internal.Sqlizer {
	buf := strings.Builder{}
	l := len(paths) - 1
	for i, path := range paths {
		buf.WriteByte('\'')
		buf.WriteString(mysqlEscape(path))
		buf.WriteByte('\'')
		if i < l {
			buf.WriteByte(',')
		}
	}
	return RAW(fmt.Sprintf("json_remove(%s,%s)", column, buf.String()))
}
