package sqls

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/snow/ini"
	"log"
	"math/rand"
	"time"
)

var master *sqlx.DB
var slaves []*sqlx.DB
var driverName string

type executor interface {
	sqlx.ExecerContext
	sqlx.QueryerContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
}

type mysqlUtilsKeyT int

const txKey mysqlUtilsKeyT = 0x10002
const justMasterKey mysqlUtilsKeyT = 0x10003

func JustUseMaster(ctx context.Context) context.Context {
	return context.WithValue(ctx, justMasterKey, true)
}

func Tx(ctx context.Context) (context.Context, func()) {
	if ctx == nil {
		panic("snow.sql: nil context")
	}

	tx := master.MustBegin()
	return context.WithValue(ctx, txKey, tx), func() {
		err := recover()
		if err == nil {
			ce := tx.Commit()
			if ce != nil {
				log.Printf("snow.sql: commit error, %s\r\n", ce.Error())
				panic(ce)
			}
			return
		}
		re := tx.Rollback()
		if re != nil {
			log.Printf("snow.sql: rollback error, %s\r\n", re.Error())
			panic(re)
		}
		panic(err)
	}
}

//noinspection GoExportedFuncWithUnexportedType
func Executor(ctx context.Context) executor {
	tx := ctx.Value(txKey)
	if tx != nil {
		return tx.(*sqlx.Tx)
	}

	if ctx.Value(justMasterKey) != nil {
		return master
	}

	if len(slaves) < 1 {
		return master
	}
	rand.Seed(time.Now().UnixNano())
	ind := rand.Int() % len(slaves)
	return slaves[ind]
}

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

func Init() {
	master = ini.SqlMaster()
	slaves = ini.SqlSlaves()
	driverName = ini.SqlDriverName()
}

func Master() *sqlx.DB {
	return master
}
