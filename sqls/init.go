package sqls

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/ini"
	"log"
	"strings"
)

var config *ini.Ini
var leader *sqlx.DB
var followers []*sqlx.DB
var sqlLog bool
var isPostgres bool
var doCreate func(ctx context.Context, op *Operator, q string, args []interface{}) int64

func Init(conf *ini.Ini, leaderV *sqlx.DB, followersV []*sqlx.DB) {
	config = conf
	leader = leaderV
	followers = followersV
	sqlLog = config.IsDebug() && config.GetBool("sql.log")
	isPostgres = leader.DriverName() == "postgres"

	doCreate = mysqlCreate
	if isPostgres {
		doCreate = postgresCreate
	}
}

func doSqlLog(q string, args []interface{}) {
	if !sqlLog {
		return
	}

	if len(args) < 1 {
		log.Printf("suna.sqlu.log: `%s`\n", q)
		return
	}

	s := fmt.Sprintf(strings.Repeat("%v,", len(args)), args...)

	log.Printf("suna.sqlu.log: `%s` `%s`\n", q, s)
}
