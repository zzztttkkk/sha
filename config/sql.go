package config

import (
	"github.com/jmoiron/sqlx"
	"math/rand"
	"time"
)

func newSqlDB(dn, url string, maxLifeTime time.Duration, openConns int) *sqlx.DB {
	db, err := sqlx.Open(dn, url)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(maxLifeTime)
	db.SetMaxOpenConns(openConns)

	if err = db.Ping(); err != nil {
		panic(err)
	}
	return db
}

func (t *Suna) SqlLeader() *sqlx.DB {
	if t.Internal.sqlLeader != nil {
		return t.Internal.sqlLeader
	}
	if len(t.Sql.Driver) < 1 {
		return nil
	}

	t.Internal.sqlLeader = newSqlDB(t.Sql.Driver, t.Sql.Leader, t.Sql.MaxLifetime.Duration, t.Sql.MaxOpen)
	return t.Internal.sqlLeader
}

func (t *Suna) SqlFollower() *sqlx.DB {
	if t.Internal.sqlFollowers != nil {
		return t.randomFollower()
	}
	if t.Internal.sqlNullFollowers || len(t.Sql.Driver) < 1 {
		return nil
	}
	for _, url := range t.Sql.Followers {
		t.Internal.sqlFollowers = append(t.Internal.sqlFollowers, newSqlDB(t.Sql.Driver, url, t.Sql.MaxLifetime.Duration, t.Sql.MaxOpen))
	}
	if len(t.Internal.sqlFollowers) < 1 {
		t.Internal.sqlNullFollowers = true
	}
	return t.randomFollower()
}

func (t *Suna) randomFollower() *sqlx.DB {
	if len(t.Internal.sqlFollowers) > 0 {
		rand.Seed(time.Now().UnixNano())
		return t.Internal.sqlFollowers[rand.Int()%len(t.Internal.sqlFollowers)]
	}
	return nil
}
