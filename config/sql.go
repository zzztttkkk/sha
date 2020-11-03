package config

import (
	"github.com/jmoiron/sqlx"
	"math/rand"
	"time"
)

func (t *Suna) newSqlDB(url string, maxLifeTime time.Duration, openConns int) *sqlx.DB {
	db := sqlx.MustConnect(t.Sql.Driver, url)
	db.SetConnMaxLifetime(maxLifeTime)
	db.SetMaxOpenConns(openConns)
	return db
}

// GetSqlLeader get sql leader
func (t *Suna) GetSqlLeader() *sqlx.DB {
	if t.internal.sqlLeader == nil {
		t.internal.sqlLeader = t.newSqlDB(t.Sql.Leader, t.Sql.MaxLifetime.Duration, t.Sql.MaxOpen)
	}
	return t.internal.sqlLeader
}

func (t *Suna) GetAnySqlFollower() *sqlx.DB {
	if len(t.internal.sqlFollowers) < 1 {
		for _, url := range t.Sql.Followers {
			t.internal.sqlFollowers = append(
				t.internal.sqlFollowers,
				t.newSqlDB(url, t.Sql.MaxLifetime.Duration, t.Sql.MaxOpen),
			)
		}
	}
	return t.randomFollower()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (t *Suna) randomFollower() *sqlx.DB {
	if len(t.internal.sqlFollowers) > 0 {
		return t.internal.sqlFollowers[rand.Int()%len(t.internal.sqlFollowers)]
	}
	return t.GetSqlLeader()
}
