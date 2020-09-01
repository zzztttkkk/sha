package config

import (
	"github.com/jmoiron/sqlx"
	"math/rand"
	"time"
)

func newSqlDB(dn, url string, maxLifeTime time.Duration, openConns int) *sqlx.DB {
	db := sqlx.MustConnect(dn, url)
	db.SetConnMaxLifetime(maxLifeTime)
	db.SetMaxOpenConns(openConns)
	return db
}

func (t *Suna) GetSqlLeader() *sqlx.DB {
	if t.Internal.sqlLeader != nil {
		return t.Internal.sqlLeader
	}
	if len(t.Sql.Driver) < 1 {
		return nil
	}

	t.Internal.sqlLeader = newSqlDB(t.Sql.Driver, t.Sql.Leader, t.Sql.MaxLifetime.Duration, t.Sql.MaxOpen)
	return t.Internal.sqlLeader
}

func (t *Suna) GetAnySqlFollower() *sqlx.DB {
	if len(t.Internal.sqlFollowers) > 0 {
		return t.randomFollower()
	}
	if len(t.Sql.Driver) < 1 {
		return nil
	}
	for _, url := range t.Sql.Followers {
		t.Internal.sqlFollowers = append(t.Internal.sqlFollowers, newSqlDB(t.Sql.Driver, url, t.Sql.MaxLifetime.Duration, t.Sql.MaxOpen))
	}
	return t.randomFollower()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (t *Suna) randomFollower() *sqlx.DB {
	if len(t.Internal.sqlFollowers) > 0 {
		return t.Internal.sqlFollowers[rand.Int()%len(t.Internal.sqlFollowers)]
	}
	return nil
}
