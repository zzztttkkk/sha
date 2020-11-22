package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"math/rand"
	"time"
)

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

var wdb *sql.DB
var rdbs []*sql.DB
var driver string

type _Key int

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	dbKey = _Key(iota)
	txKey
	justWKey
	driverKey
)

func UseWriteableDB(ctx context.Context) context.Context {
	v := getExecutor(ctx)
	if v != nil {
		panic(fmt.Errorf("suna.sqlx: Executor is not nil"))
	}
	return context.WithValue(ctx, justWKey, 1)
}

func getExecutor(ctx context.Context) interface{} {
	var v interface{}
	v = ctx.Value(txKey)
	if v != nil {
		return v
	}
	v = ctx.Value(dbKey)
	if v != nil {
		return v
	}
	return nil
}

func ExcScanner(ctx context.Context) *_ExcScanner {
	v := getExecutor(ctx)
	if v != nil {
		return &_ExcScanner{executor: v.(Executor)}
	}
	if ctx.Value(justWKey) != nil {
		return &_ExcScanner{executor: wdb}
	}
	if len(rdbs) < 1 {
		return &_ExcScanner{executor: wdb}
	}
	return &_ExcScanner{executor: rdbs[rand.Int()%len(rdbs)]}
}

func Tx(ctx context.Context, opts *sql.TxOptions) (context.Context, func()) {
	v := ctx.Value(txKey)
	if v != nil {
		panic(errors.New("suna.sqlx: sub tx is invalid"))
	}

	tx, err := wdb.BeginTx(ctx, opts)
	if err != nil {
		panic(err)
	}
	return context.WithValue(
			context.WithValue(ctx, driverKey, wdb.Driver()), txKey, tx,
		), func() {
			recv := recover()
			if recv == nil {
				commitErr := tx.Commit()
				if commitErr != nil {
					internal.L.Printf("suna.sqlx: commit error, %s\r\n", commitErr.Error())
					panic(commitErr)
				}
				return
			}
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				internal.L.Printf("suna.sqlx: rollback error: %s\r\n", rollbackErr.Error())
				panic(recv)
			}
			panic(recv)
		}
}
