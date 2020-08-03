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

func (t *Type) SqlLeader() *sqlx.DB {
	if t.Sql.l != nil {
		return t.Sql.l
	}
	if len(t.Sql.Driver) < 1 {
		return nil
	}

	t.Sql.l = newSqlDB(t.Sql.Driver, t.Sql.Leader, t.Sql.MaxLifetime, t.Sql.MaxOpen)
	return t.Sql.l
}

func (t *Type) SqlFollower() *sqlx.DB {
	if t.Sql.fs != nil {
		rand.Seed(time.Now().UnixNano())
		return t.Sql.fs[rand.Int()%len(t.Sql.fs)]
	}
	if t.Sql.nfs || len(t.Sql.Driver) < 1 {
		return nil
	}
	for _, url := range t.Sql.Followers {
		t.Sql.fs = append(t.Sql.fs, newSqlDB(t.Sql.Driver, url, t.Sql.MaxLifetime, t.Sql.MaxOpen))
	}
	if len(t.Sql.fs) < 1 {
		t.Sql.nfs = true
	}
	return t.SqlFollower()
}
