package sqls

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"log"
	"strings"
)

var cfg *config.Config
var isPostgres bool
var doCreate func(ctx context.Context, op *Operator, q string, args []interface{}) int64
var leader *sqlx.DB

func doSqlLog(q string, args []interface{}) {
	if !cfg.Sql.Log {
		return
	}

	if len(args) < 1 {
		log.Printf("suna.sqls.log: `%s`\n", q)
		return
	}

	s := fmt.Sprintf(strings.Repeat("%v,", len(args)), args...)

	log.Printf("suna.sqls.log: `%s` `%s`\n", q, s)
}

func init() {
	internal.LazyInvoke(
		func(conf *config.Config) {
			cfg = conf
			if cfg.SqlLeader() == nil {
				log.Println("suna.sqls: init error")
				return
			}
			leader = cfg.SqlLeader()
			isPostgres = leader.DriverName() == "postgres"
			doCreate = mysqlCreate
			if isPostgres {
				doCreate = postgresCreate
			}
		},
	)
}
