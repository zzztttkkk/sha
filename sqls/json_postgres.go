// +build postgres

package sqls

import (
	"bytes"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/savsgio/gotils"
	"github.com/zzztttkkk/suna/jsonx"
	"github.com/zzztttkkk/suna/sqls/internal"
)

func init() {
	driver = "postgres"
	_JsonSetImpl = json_set
	_JsonUpdateImpl = json_update
	_JsonRemoveImpl = json_remove
}

var (
	qs   = "'"
	eqs  = "''"
	qsb  = []byte(qs)
	eqsb = []byte(eqs)

	dqs   = "\""
	edqs  = "\\\""
	dqsb  = []byte(qs)
	edqsb = []byte(eqs)
)

func escapeSingleQuote(s string) string { return strings.ReplaceAll(s, qs, eqs) }

func escapeSingleQuoteBytes(s []byte) []byte { return bytes.ReplaceAll(s, qsb, eqsb) }

func escapeDoubleQuote(s string) string { return strings.ReplaceAll(s, dqs, edqs) }

func escapeDoubleQuoteBytes(s []byte) []byte { return bytes.ReplaceAll(s, dqsb, edqsb) }

func json_set(column string, path string, v interface{}) internal.Sqlizer {
	return RAW(
		fmt.Sprintf(
			"jsonb_set(%s::jsonb, '{\"%s\"}', '%s')",
			column,
			escapeDoubleQuote(path),
			gotils.B2S(escapeSingleQuoteBytes(jsonx.MustMarshal(v))),
		),
	)
}

func json_update(column string, m map[string]interface{}) internal.Sqlizer {
	return RAW(
		fmt.Sprintf(
			"%s::jsonb||'%s'::jsonb",
			column,
			gotils.B2S(escapeSingleQuoteBytes(jsonx.MustMarshal(m))),
		),
	)
}

func json_remove(column string, paths ...string) internal.Sqlizer {
	buf := strings.Builder{}
	buf.WriteString(column)
	buf.WriteString("::jsonb")
	for _, path := range paths {
		buf.WriteByte('-')
		buf.WriteByte('\'')
		buf.WriteString(escapeSingleQuote(path))
		buf.WriteByte('\'')
	}
	return RAW(buf.String())
}
