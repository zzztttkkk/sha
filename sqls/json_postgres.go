package sqls

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/savsgio/gotils"
	"github.com/zzztttkkk/suna/jsonx"
	"github.com/zzztttkkk/suna/sqls/internal"
)

var (
	postgresQs   = "'"
	postgresEqs  = "''"
	postgresQsb  = []byte(postgresQs)
	postgresEqsb = []byte(postgresEqs)

	postgresDqs  = "\""
	postgresEdqs = "\\\""
)

func escapeSingleQuote(s string) string { return strings.ReplaceAll(s, postgresQs, postgresEqs) }

func escapeSingleQuoteBytes(s []byte) []byte { return bytes.ReplaceAll(s, postgresQsb, postgresEqsb) }

func escapeDoubleQuote(s string) string { return strings.ReplaceAll(s, postgresDqs, postgresEdqs) }

func postgresJsonSet(column string, path string, v interface{}) internal.Sqlizer {
	return RAW(
		fmt.Sprintf(
			"jsonb_set(%s::jsonb, '{\"%s\"}', '%s')",
			column,
			escapeDoubleQuote(path),
			gotils.B2S(escapeSingleQuoteBytes(jsonx.MustMarshal(v))),
		),
	)
}

func postgresJsonUpdate(column string, m map[string]interface{}) internal.Sqlizer {
	return RAW(
		fmt.Sprintf(
			"%s::jsonb||'%s'::jsonb",
			column,
			gotils.B2S(escapeSingleQuoteBytes(jsonx.MustMarshal(m))),
		),
	)
}

func postgresJsonRemove(column string, paths ...string) internal.Sqlizer {
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
