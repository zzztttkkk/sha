package sqlu

import (
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

func Init(conf *ini.Ini, leaderV *sqlx.DB, followersV []*sqlx.DB) {
	config = conf
	leader = leaderV
	followers = followersV
	sqlLog = config.IsDebug() && config.GetBool("sql.log")
}

func doSqlLog(q string, args []interface{}) {
	if len(args) < 1 {
		log.Printf("suna.sqlu.log: `%s`\n", q)
		return
	}

	s := fmt.Sprintf(strings.Repeat("%v,", len(args)), args...)

	log.Printf("suna.sqlu.log: `%s` `%s`\n", q, s)
}
