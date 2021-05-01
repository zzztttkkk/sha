package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	x "github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/sha/utils"
	mrandlib "math/rand"
	"sync"
)

var dbLock sync.Mutex
var writableDb *x.DB
var readonlyDbs []*x.DB

func OpenWriteableDB(driverName, uri string) {
	dbLock.Lock()
	defer dbLock.Unlock()

	if writableDb != nil {
		return
	}
	writableDb = x.MustOpen(driverName, uri)
}

func OpenReadableDB(drivername, uri string) {
	dbLock.Lock()
	defer dbLock.Unlock()

	readonlyDbs = append(readonlyDbs, x.MustOpen(drivername, uri))
}

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
	justWDBKey
)

func UseWriteableDB(ctx context.Context) context.Context {
	return context.WithValue(ctx, justWDBKey, true)
}

type TxOptions struct {
	sql.TxOptions
	SavePointName string
}

type TxWrapper struct {
	LoggingExecutor
	savepoint string
}

func (t *TxWrapper) Savepoint() string { return t.savepoint }

func (t *TxWrapper) Commit(ctx context.Context) error {
	if len(t.savepoint) > 0 {
		if e := t.LoggingExecutor.releaseSavePoint(ctx, fmt.Sprintf("%s_BEGIN", t.savepoint)); e != nil {
			return e
		}
		return t.LoggingExecutor.savePoint(ctx, t.savepoint)
	}
	return t.LoggingExecutor.Executor.(*x.Tx).Commit()
}

func (t *TxWrapper) Rollback(ctx context.Context) error {
	if len(t.savepoint) > 0 {
		return t.LoggingExecutor.rollbackToSavePoint(ctx, fmt.Sprintf("%s_BEGIN", t.savepoint))
	}
	return t.LoggingExecutor.Executor.(*x.Tx).Rollback()
}

type AutoCommitError struct {
	v interface{}
	e error
}

func (r *AutoCommitError) Recoverd() interface{} { return r.v }

func (r AutoCommitError) DbError() error { return r.e }

func (r *AutoCommitError) Error() string {
	return fmt.Sprintf("sha.sqlx: database error `%v`, recoverd value `%v`", r.e.Error(), r.v)
}

func (t *TxWrapper) AutoCommit(ctx context.Context) {
	v := recover()
	var e error
	if v == nil {
		e = t.Commit(ctx)
	} else {
		e = t.Rollback(ctx)
	}
	if e != nil {
		panic(&AutoCommitError{v: v, e: e})
	}
}

func TxWithOptions(ctx context.Context, options *TxOptions) (context.Context, *TxWrapper) {
	txi := ctx.Value(txKey)
	var tx *x.Tx
	if txi != nil { // sub tx
		if options == nil || options.SavePointName == "" {
			panic(errors.New("sha.sqlx: empty options or empty savepoint name"))
		}

		switch tv := txi.(type) {
		case *x.Tx:
			tx = tv
		case *TxWrapper:
			tx = tv.Executor.(*x.Tx)
		}

		subTx := &TxWrapper{LoggingExecutor: LoggingExecutor{Executor: tx}, savepoint: options.SavePointName}
		if e := subTx.LoggingExecutor.savePoint(ctx, fmt.Sprintf("%s_BEGIN", subTx.savepoint)); e != nil {
			panic(e)
		}
		return context.WithValue(ctx, txKey, subTx), subTx
	}

	var o *sql.TxOptions
	if options != nil {
		o = &options.TxOptions
	}
	tx = writableDb.MustBeginTx(ctx, o)
	return context.WithValue(ctx, txKey, tx), &TxWrapper{LoggingExecutor: LoggingExecutor{Executor: tx}}
}

func Tx(ctx context.Context) (context.Context, *TxWrapper) { return TxWithOptions(ctx, nil) }

func init() {
	utils.MathRandSeed()
}

var PickReadonlyDB = func(dbs []*x.DB) *x.DB { return dbs[int(mrandlib.Uint32())%len(dbs)] }

func Exe(ctx context.Context) LoggingExecutor {
	tx := ctx.Value(txKey)
	if tx != nil {
		return LoggingExecutor{tx.(*x.Tx)}
	}

	if ctx.Value(justWDBKey) != nil {
		return LoggingExecutor{writableDb}
	}
	if len(readonlyDbs) < 1 {
		return LoggingExecutor{writableDb}
	}
	return LoggingExecutor{PickReadonlyDB(readonlyDbs)}
}

func db(ctx context.Context) *x.DB {
	exe := Exe(ctx).Executor
	if d, ok := exe.(*x.DB); ok {
		return d
	}
	// tx
	return writableDb
}
