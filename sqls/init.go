package sqls

import (
	"fmt"
	"log"
	"strings"

	"github.com/zzztttkkk/suna/config"
	si "github.com/zzztttkkk/suna/internal"
	ci "github.com/zzztttkkk/suna/sqls/internal"
)

var cfg *config.Suna
var builder *ci.Builder

func _DoSqlLogging(q string, args []interface{}) {
	if !cfg.Sql.Logging {
		return
	}

	if len(args) < 1 {
		log.Printf("suna.sqls.log: `%s`\n", q)
		return
	}

	for i, v := range args {
		switch rv := v.(type) {
		case []byte:
			args[i] = string(rv)
		}
	}

	s := fmt.Sprintf(strings.Repeat("'%v',", len(args)), args...)
	log.Printf("suna.sqls.log: `%s` with args `[%s]`\n", q, s)
}

func init() {
	si.Dig.LazyInvoke(
		func(conf *config.Suna) {
			cfg = conf
			if cfg.GetSqlLeader() == nil {
				log.Println("suna.sqls: init error")
				return
			}
			builder = ci.NewBuilder(cfg.GetSqlLeader())
		},
	)
}

func Select(columns ...string) *ci.SelectBuilder {
	return builder.Select(columns...)
}

func Insert(into string) *ci.InsertBuilder {
	return builder.Insert(into)
}

func Update(table string) *ci.UpdateBuilder {
	return builder.Update(table)
}

func Delete(what ...string) *ci.DeleteBuilder {
	return builder.Delete(what...)
}

func IsPostgres() bool {
	return builder.IsPostgres()
}
