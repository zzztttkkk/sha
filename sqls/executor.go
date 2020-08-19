package sqls

import (
	"context"
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
}

type ctxKeyT int

const (
	txKey = ctxKeyT(iota + 1000)
	userKey
	justLeaderKey
)

func JustUseLeader(ctx context.Context) context.Context {
	return context.WithValue(ctx, justLeaderKey, true)
}

func doNothing() {}

func TxByUser(ctx *fasthttp.RequestCtx) (context.Context, func()) {
	return Tx(context.WithValue(ctx, userKey, auth.MustGetUser(ctx)))
}

// starts a transaction, return a sub context and a commit function
func Tx(ctx context.Context) (context.Context, func()) {
	_tx := ctx.Value(txKey)
	if _tx != nil {
		return ctx, doNothing
	}

	tx := leader.MustBegin()

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

func TxOperator(ctx context.Context) auth.User {
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
		return leader
	}

	f := cfg.SqlFollower()
	if f != nil {
		return f
	}
	return cfg.SqlLeader()
}
