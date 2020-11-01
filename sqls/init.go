package sqls

import (
	"fmt"
	"log"
	"strings"

	"github.com/zzztttkkk/suna/config"
	SunaInternal "github.com/zzztttkkk/suna/internal"
	SqlsInternal "github.com/zzztttkkk/suna/sqls/internal"
)

var cfg *config.Suna
var builder *SqlsInternal.Builder

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
	log.Printf("suna.sqls.log: `%s` with args [%s]\n", q, s)
}

func initMysqlJson() {
	jsonSetImpl = mysqlJsonSet
	jsonUpdateImpl = mysqlJsonUpdate
	jsonRemoveImpl = mysqlJsonRemove
}

func initPostgresJson() {
	jsonSetImpl = postgresJsonSet
	jsonUpdateImpl = postgresJsonUpdate
	jsonRemoveImpl = postgresJsonRemove
}

func init() {
	SunaInternal.Dig.Append(
		func(conf *config.Suna) {
			cfg = conf
			if cfg.GetSqlLeader() == nil {
				log.Println("suna.sqls: init error")
				return
			}
			builder = SqlsInternal.NewBuilder(cfg.GetSqlLeader())

			if builder.IsPostgres() {
				initPostgresJson()
			} else {
				initMysqlJson()
			}
		},
	)
}

func Select(columns ...string) *SqlsInternal.SelectBuilder {
	return builder.Select(columns...)
}

func Insert(columns ...string) *SqlsInternal.InsertBuilder {
	return builder.Insert().Columns(columns...)
}

func Update() *SqlsInternal.UpdateBuilder {
	return builder.Update()
}

func Delete() *SqlsInternal.DeleteBuilder {
	return builder.Delete()
}

func IsPostgres() bool {
	return builder.IsPostgres()
}
