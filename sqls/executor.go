package sqls

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"

	"github.com/zzztttkkk/snow/ini"
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

var config *ini.Config

func Tx(ctx context.Context) (context.Context, func()) {
	if ctx == nil {
		panic("snow.sql: nil context")
	}

	tx := config.SqlLeader().MustBegin()
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
		return config.SqlLeader()
	}

	db := config.SqlFollower()
	if db == nil {
		return config.SqlLeader()
	}
	return db
}

func Init(conf *ini.Config) {
	config = conf
}
