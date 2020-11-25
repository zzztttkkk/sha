package sqlx

import (
	"context"
	"errors"
	x "github.com/jmoiron/sqlx"
	"math/rand"
	"time"
)

var wdb *x.DB
var rdbs []*x.DB

type Executor interface {
	x.ExecerContext
	x.QueryerContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	PreparexContext(ctx context.Context, query string) (*x.Stmt, error)
	DriverName() string
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
}

type _Key int

const (
	txKey = _Key(iota + 1000)
	justLeaderKey
)

func WriteableDB(ctx context.Context) context.Context {
	return context.WithValue(ctx, justLeaderKey, true)
}

// starts a transaction, return a sub context and a commit function
func Tx(ctx context.Context) (nctx context.Context, committer func()) {
	_tx := ctx.Value(txKey)
	if _tx != nil {
		panic(errors.New("suna.sqlx: sub tx is invalid"))
	}

	tx := wdb.MustBegin()
	return context.WithValue(ctx, txKey, tx), func() {
		recv := recover()
		if recv == nil {
			commitErr := tx.Commit()
			if commitErr != nil {
				panic(commitErr)
			}
			return
		}
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			panic(rollbackErr)
		}
		panic(recv)
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var PickReadonlyDB = func(dbs []*x.DB) *x.DB { return dbs[rand.Int()%len(dbs)] }

func Exe(ctx context.Context) W {
	tx := ctx.Value(txKey)
	if tx != nil {
		return W{tx.(*x.Tx)}
	}
	if ctx.Value(justLeaderKey) != nil {
		return W{wdb}
	}
	if len(rdbs) < 1 {
		return W{wdb}
	}
	return W{PickReadonlyDB(rdbs)}
}

func db(ctx context.Context) *x.DB {
	tx := ctx.Value(txKey)
	if tx != nil {
		return wdb
	}
	if ctx.Value(justLeaderKey) != nil {
		return wdb
	}
	if len(rdbs) < 1 {
		return wdb
	}
	return PickReadonlyDB(rdbs)
}
