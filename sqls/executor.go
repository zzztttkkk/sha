package sqls

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
)

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
	tx := leader.MustBegin()

	return context.WithValue(ctx, txKey, tx), func() {
		err := recover()
		if err == nil {
			ce := tx.Commit()
			if ce != nil {
				log.Printf("suna.sql: commit error, %s\r\n", ce.Error())
				panic(ce)
			}
			return
		}
		re := tx.Rollback()
		if re != nil {
			log.Printf("suna.sql: rollback error, %s\r\n", re.Error())
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
		return leader
	}

	if len(followers) < 1 {
		return leader
	}
	return Follower()
}

func Leader() *sqlx.DB { return leader }

func Follower() *sqlx.DB {
	rand.Seed(time.Now().UnixNano())
	return followers[rand.Int()%len(followers)]
}
