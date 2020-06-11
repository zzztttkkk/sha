package ini

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

var (
	master *sqlx.DB
	slaves []*sqlx.DB
)

func newSqlDB(url string, maxLifeTime time.Duration, maxOpenConns int) *sqlx.DB {
	db, err := sqlx.Open("mysql", url)
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

func initSqls() {
	maxLifetime := time.Second * time.Duration(GetIntOr("mysql.max_life_time", 7200))
	maxOpenConns := int(GetIntOr("mysql.max_open_num", 5))

	master = newSqlDB(string(GetMust("mysql.master")), maxLifetime, maxOpenConns)

	for i := 0; i < int(GetIntOr("mysql.slaves", 0)); i++ {
		slaves = append(slaves, newSqlDB(string(GetMust(fmt.Sprintf("mysql.slave.%d", i))), maxLifetime, maxOpenConns))
	}
}

func SqlMaster() *sqlx.DB {
	return master
}

func SqlSlaves() []*sqlx.DB {
	return slaves
}
