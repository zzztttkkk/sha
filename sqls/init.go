package sqls

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"log"
	"math/rand"
	"strings"
	"time"
)

var cfg *config.Suna

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
	internal.Dig.LazyInvoke(
		func(conf *config.Suna) {
			cfg = conf
			if cfg.GetSqlLeader() == nil {
				log.Println("suna.sqls: init error")
				return
			}
		},
	)
}

type _Dbs struct {
	Leader    *sqlx.DB
	Followers []*sqlx.DB
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (dbs *_Dbs) _RandomFollower() *sqlx.DB {
	if len(dbs.Followers) < 1 {
		return nil
	}
	return dbs.Followers[rand.Int()%len(dbs.Followers)]
}

var _DbGroups = map[string]*_Dbs{}

func AddDBGroup(name, driverName, leader string, followers []string, maxLifeTime time.Duration, openConns int) {
	if len(name) < 1 {
		panic(fmt.Errorf("suna.sqls: empty name"))
	}
	if _, ok := _DbGroups[name]; ok {
		panic(fmt.Errorf("suna.sqls: database group `%s` is exists\n", name))
	}

	dbs := &_Dbs{}
	dbs.Leader = newSqlDB(driverName, leader, maxLifeTime, openConns)
	for _, f := range followers {
		dbs.Followers = append(dbs.Followers, newSqlDB(driverName, f, maxLifeTime, openConns))
	}
	_DbGroups[name] = dbs
}

func newSqlDB(dn, url string, maxLifeTime time.Duration, openConns int) *sqlx.DB {
	db := sqlx.MustConnect(dn, url)
	db.SetConnMaxLifetime(maxLifeTime)
	db.SetMaxOpenConns(openConns)
	return db
}
