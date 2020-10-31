package config

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"math/rand"
	"time"
)

func newSqlDB(url string, maxLifeTime time.Duration, openConns int) *sqlx.DB {
	db := sqlx.MustConnect(sql.Drivers()[0], url)
	db.SetConnMaxLifetime(maxLifeTime)
	db.SetMaxOpenConns(openConns)
	return db
}

// GetSqlLeader get sql leader
func (t *Suna) GetSqlLeader() *sqlx.DB {
	if t.Internal.sqlLeader != nil {
		return t.Internal.sqlLeader
	}

	t.Internal.sqlLeader = newSqlDB(t.Sql.Leader, t.Sql.MaxLifetime.Duration, t.Sql.MaxOpen)
	return t.Internal.sqlLeader
}

func (t *Suna) GetAnySqlFollower() *sqlx.DB {
	if len(t.Internal.sqlFollowers) > 0 {
		return t.randomFollower()
	}
	for _, url := range t.Sql.Followers {
		t.Internal.sqlFollowers = append(t.Internal.sqlFollowers, newSqlDB(url, t.Sql.MaxLifetime.Duration, t.Sql.MaxOpen))
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
