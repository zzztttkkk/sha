package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	x "github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/sha/utils"
	mrandlib "math/rand"
	"regexp"
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
	subTxKey
)

func UseWriteableDB(ctx context.Context) context.Context {
	return context.WithValue(ctx, justWDBKey, true)
}

type TxOptions struct {
	sql.TxOptions
	SavePointName string
}

var savepointNameRegexp = regexp.MustCompile(`^[a-zA-Z_]\w*$`)

func ensureSavePointName(name string) string {
	if !savepointNameRegexp.MatchString(name) {
		panic(fmt.Errorf("sha.sqlx: bad savepoint name `%s`", name))
	}
	return name
}

func savePoint(ctx context.Context, tx *x.Tx, name string) error {
	_, e := _LoggingWrapper{Executor: tx}.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", ensureSavePointName(name)), nil)
	return e
}

func rollbackToSavePoint(ctx context.Context, tx *x.Tx, name string) error {
	_, e := _LoggingWrapper{Executor: tx}.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", ensureSavePointName(name)), nil)
	return e
}

func releaseSavePoint(ctx context.Context, tx *x.Tx, name string) error {
	_, e := _LoggingWrapper{Executor: tx}.Exec(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", ensureSavePointName(name)), nil)
	return e
}

const subTxBegin = "SHA_SQLX_SUB_TX_BEGIN"

func getSubTx(ctx context.Context) *TxWrapper {
	i := ctx.Value(subTxKey)
	if i == nil {
		return nil
	}
	return i.(*TxWrapper)
}

type TxWrapper struct {
	_LoggingWrapper
	savepoint string
	done      bool
}

func (t *TxWrapper) Savepoint() string { return t.savepoint }

func (t *TxWrapper) Commit(ctx context.Context) error {
	if len(t.savepoint) > 0 {
		if e := t._LoggingWrapper.ReleaseSavePoint(ctx, subTxBegin); e != nil {
			return e
		}
		t.done = true
		return t._LoggingWrapper.SavePoint(ctx, t.savepoint)
	}
	return t._LoggingWrapper.Executor.(*x.Tx).Commit()
}

func (t *TxWrapper) Rollback(ctx context.Context) error {
	if len(t.savepoint) > 0 {
		t.done = true
		return t._LoggingWrapper.RollbackToSavePoint(ctx, subTxBegin)
	}
	return t._LoggingWrapper.Executor.(*x.Tx).Rollback()
}

// TxWithOptions starts a transaction, return a sub context and a commit function
func TxWithOptions(ctx context.Context, options *TxOptions) (context.Context, *TxWrapper) {
	var tx *x.Tx
	if ctx.Value(txKey) != nil { // sub tx
		if options == nil || options.SavePointName == "" {
			panic(errors.New("sha.sqlx: empty options or empty savepoint name"))
		}

		tx = ctx.Value(txKey).(*x.Tx)
		subTx := getSubTx(ctx)
		if subTx != nil && !subTx.done {
			panic(fmt.Errorf("sha.sqlx: unfinished sub tx: `%s`", subTx.savepoint))
		}
		subTx = &TxWrapper{_LoggingWrapper: _LoggingWrapper{Executor: tx}, savepoint: options.SavePointName}
		if e := subTx._LoggingWrapper.SavePoint(ctx, subTxBegin); e != nil {
			panic(e)
		}
		return context.WithValue(ctx, subTxKey, subTx), subTx
	}

	var o *sql.TxOptions
	if options != nil {
		o = &options.TxOptions
	}
	tx = writableDb.MustBeginTx(ctx, o)
	return context.WithValue(ctx, txKey, tx), &TxWrapper{_LoggingWrapper: _LoggingWrapper{Executor: tx}}
}

func Tx(ctx context.Context) (context.Context, *TxWrapper) { return TxWithOptions(ctx, nil) }

func init() {
	utils.MathRandSeed()
}

var PickReadonlyDB = func(dbs []*x.DB) *x.DB { return dbs[int(mrandlib.Uint32())%len(dbs)] }

func Exe(ctx context.Context) _LoggingWrapper {
	tx := ctx.Value(txKey)
	if tx != nil {
		return _LoggingWrapper{tx.(*x.Tx)}
	}

	if ctx.Value(justWDBKey) != nil {
		return _LoggingWrapper{writableDb}
	}
	if len(readonlyDbs) < 1 {
		return _LoggingWrapper{writableDb}
	}
	return _LoggingWrapper{PickReadonlyDB(readonlyDbs)}
}

func db(ctx context.Context) *x.DB {
	exe := Exe(ctx).Executor
	if d, ok := exe.(*x.DB); ok {
		return d
	}
	// tx
	return writableDb
}
