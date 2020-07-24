package ini

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

func newSqlDB(dn, url string, maxLifeTime time.Duration, maxOpenConns int) *sqlx.DB {
	db, err := sqlx.Open(dn, url)
	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(maxLifeTime)
	db.SetMaxOpenConns(maxOpenConns)

	if err = db.Ping(); err != nil {
		panic(err)
	}
	return db
}

func (conf *Ini) SqlClients() (*sqlx.DB, []*sqlx.DB) {
	if conf.sqlL != nil {
		return conf.sqlL, conf.sqlF
	}

	maxLifetime := time.Second * time.Duration(conf.GetIntOr("sql.max_life_time", 7200))
	maxOpenConns := int(conf.GetIntOr("sql.max_open_num", 5))
	dn := string(conf.GetMust("sql.driver"))

	conf.sqlL = newSqlDB(dn, string(conf.GetMust("sql.leader")), maxLifetime, maxOpenConns)

	for i := 0; i < int(conf.GetIntOr("sql.followers", 0)); i++ {
		conf.sqlF = append(conf.sqlF, newSqlDB(dn, string(conf.GetMust(fmt.Sprintf("mysql.slave.%d", i))), maxLifetime, maxOpenConns))
	}
	return conf.sqlL, conf.sqlF
}
