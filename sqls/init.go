package sqls

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"log"
	"strings"
)

var cfg *config.Suna
var isPostgres bool
var leader *sqlx.DB

func doSqlLog(q string, args []interface{}) {
	if !cfg.Sql.Log {
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

	log.Printf("suna.sqls.log: `%s` [%s]\n", q, s)
}

func init() {
	internal.Dig.LazyInvoke(
		func(conf *config.Suna) {
			cfg = conf
			if cfg.SqlLeader() == nil {
				log.Println("suna.sqls: init error")
				return
			}
			leader = cfg.SqlLeader()
			isPostgres = leader.DriverName() == "postgres"
		},
	)
}
