package ini

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

var (
	master     *sqlx.DB
	slaves     []*sqlx.DB
	driverName string
)

func newSqlDB(url string, maxLifeTime time.Duration, maxOpenConns int) *sqlx.DB {
	db, err := sqlx.Open(driverName, url)
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
	driverName = string(GetMust("sql.driver"))
	maxLifetime := time.Second * time.Duration(GetIntOr("sql.max_life_time", 7200))
	maxOpenConns := int(GetIntOr("sql.max_open_num", 5))

	master = newSqlDB(string(GetMust("sql.master.url")), maxLifetime, maxOpenConns)

	for i := 0; i < int(GetIntOr("sql.slaves", 0)); i++ {
		slaves = append(slaves, newSqlDB(string(GetMust(fmt.Sprintf("sql.slave.%d", i))), maxLifetime, maxOpenConns))
	}
}

func SqlMaster() *sqlx.DB {
	return master
}

func SqlSlaves() []*sqlx.DB {
	return slaves
}

func SqlDriverName() string {
	return driverName
}
