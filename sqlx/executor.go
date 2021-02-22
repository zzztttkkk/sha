package sqlx

import (
	"context"
	"errors"
	"fmt"
	x "github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/sha/utils"
	mrandlib "math/rand"
)

var writableDb *x.DB
var readonlyDbs []*x.DB

func OpenWriteableDB(driverName, uri string) {
	if writableDb != nil {
		return
	}
	writableDb = x.MustOpen(driverName, uri)
}

func OpenReadableDB(drivername, uri string) {
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

type RollbackError struct {
	RecoverVal interface{}
	Err        error
}

func (re *RollbackError) Error() string {
	return fmt.Sprintf("sha.sqlx: rollback failed, %s %v", re.Err, re.RecoverVal)
}

var ErrSubTx = errors.New("sha.sqlx: sub tx is invalid")

// starts a transaction, return a sub context and a commit function
func Tx(ctx context.Context) (nctx context.Context, committer func()) {
	if ctx.Value(txKey) != nil {
		panic(ErrSubTx)
	}

	tx := writableDb.MustBegin()
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
			panic(&RollbackError{Err: rollbackErr, RecoverVal: recv})
		}
		panic(recv)
	}
}

func init() {
	utils.MathRandSeed()
}

var PickReadonlyDB = func(dbs []*x.DB) *x.DB { return dbs[int(mrandlib.Uint32())%len(dbs)] }

func Exe(ctx context.Context) W {
	tx := ctx.Value(txKey)
	if tx != nil {
		return W{tx.(*x.Tx)}
	}
	if ctx.Value(justWDBKey) != nil {
		return W{writableDb}
	}
	if len(readonlyDbs) < 1 {
		return W{writableDb}
	}
	return W{PickReadonlyDB(readonlyDbs)}
}

func db(ctx context.Context) *x.DB {
	exe := Exe(ctx).Raw
	if d, ok := exe.(*x.DB); ok {
		return d
	}
	// tx
	return writableDb
}
