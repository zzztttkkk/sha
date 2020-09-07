package sqlx

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/auth"
	"log"
)

type executor interface {
	sqlx.ExecerContext
	sqlx.QueryerContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	DriverName() string
}

type ctxKeyT int

const (
	txKey = ctxKeyT(iota + 1000)
	userKey
	justLeaderKey
	dbKey
)

func UseDBGroup(ctx context.Context, name string) context.Context {
	if len(name) < 1 {
		panic(fmt.Errorf("suna.sqls: empty name"))
	}
	return context.WithValue(ctx, dbKey, name)
}

func GetLeaderDB(ctx context.Context) *sqlx.DB {
	nameI := ctx.Value(dbKey)
	if nameI == nil {
		return cfg.GetSqlLeader()
	}
	var name string
	var ok bool
	if name, ok = nameI.(string); !ok {
		panic(fmt.Errorf("suna.sqls: error db value"))
	}
	dbs, ok := _DbGroups[name]
	if !ok {
		panic(fmt.Errorf("suna.sqls: database group `%s` is not exists", name))
	}
	return dbs.Leader
}

func GetAnyFollowerDB(ctx context.Context) *sqlx.DB {
	nameI := ctx.Value(dbKey)
	if nameI == nil {
		return cfg.GetAnySqlFollower()
	}
	var name string
	var ok bool
	if name, ok = nameI.(string); !ok {
		panic(fmt.Errorf("suna.sqls: error db value"))
	}
	dbs, ok := _DbGroups[name]
	if !ok {
		panic(fmt.Errorf("suna.sqls: database group `%s` is not exists", name))
	}
	return dbs._RandomFollower()
}

func JustUseLeaderDB(ctx context.Context) context.Context {
	return context.WithValue(ctx, justLeaderKey, true)
}

func TxByUser(ctx *fasthttp.RequestCtx) (context.Context, func()) {
	return Tx(context.WithValue(ctx, userKey, auth.MustGetUser(ctx)))
}

// starts a transaction, return a sub context and a commit function
func Tx(ctx context.Context) (context.Context, func()) {
	_tx := ctx.Value(txKey)
	if _tx != nil {
		panic(errors.New("suna.sqls: sub tx is invalid"))
	}

	tx := GetLeaderDB(ctx).MustBegin()
	return context.WithValue(ctx, txKey, tx), func() {
		err := recover()
		if err == nil {
			ce := tx.Commit()
			if ce != nil {
				log.Printf("suna.sqls: commit error, %s\r\n", ce.Error())
				panic(ce)
			}
			return
		}
		re := tx.Rollback()
		if re != nil {
			log.Printf("suna.sqls: rollback error, %s\r\n", re.Error())
			panic(re)
		}
		panic(err)
	}
}

func GetTxOperator(ctx context.Context) auth.User {
	u, ok := ctx.Value(userKey).(auth.User)
	if ok {
		return u
	}
	return nil
}

func Executor(ctx context.Context) executor {
	tx := ctx.Value(txKey)
	if tx != nil {
		return tx.(*sqlx.Tx)
	}
	if ctx.Value(justLeaderKey) != nil {
		return GetLeaderDB(ctx)
	}
	f := GetAnyFollowerDB(ctx)
	if f == nil {
		return GetLeaderDB(ctx)
	}
	return f
}
