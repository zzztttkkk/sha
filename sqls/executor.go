package sqls

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/auth"
	"log"
	"reflect"

	ci "github.com/zzztttkkk/suna/sqls/internal"
)

type SqlExecutor interface {
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
)

func UseLeader(ctx context.Context) context.Context {
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

	tx := cfg.GetSqlLeader().MustBegin()
	return context.WithValue(ctx, txKey, tx), func() {
		recv := recover()
		if recv == nil {
			commitErr := tx.Commit()
			if commitErr != nil {
				log.Printf("suna.sqls: commit error, %s\r\n", commitErr.Error())
				panic(commitErr)
			}
			return
		}
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Printf("suna.sqls: rollback error: %s\r\n", rollbackErr.Error())
			panic(recv)
		}
		panic(recv)
	}
}

func CurrentUser(ctx context.Context) auth.User {
	u, _ := ctx.Value(userKey).(auth.User)
	return u
}

func Executor(ctx context.Context) SqlExecutor {
	tx := ctx.Value(txKey)
	if tx != nil {
		return tx.(*sqlx.Tx)
	}
	if ctx.Value(justLeaderKey) != nil {
		return cfg.GetSqlLeader()
	}
	f := cfg.GetAnySqlFollower()
	if f == nil {
		return cfg.GetSqlLeader()
	}
	return f
}

func inTx(ctx context.Context) bool {
	return ctx.Value(txKey) != nil
}

func ExecuteCustomScan(ctx context.Context, scanner *Scanner, builder *ci.SelectBuilder) int {
	query, args, err := builder.ToSql()
	if err != nil {
		panic(err)
	}
	_DoSqlLogging(query, args)
	rows, err := Executor(ctx).QueryxContext(ctx, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		panic(err)
	}
	defer rows.Close()

	return scanner.Scan(rows)
}

func ExecuteSelectBuilder(ctx context.Context, dist interface{}, builder *ci.SelectBuilder) bool {
	q, a, e := builder.ToSql()
	return ExecuteSelect(ctx, dist, q, a, e)
}

func ExecuteSelect(ctx context.Context, dist interface{}, query string, args []interface{}, err error) bool {
	if err != nil {
		panic(err)
	}
	dT := reflect.TypeOf(dist)
	if dT.Kind() != reflect.Ptr {
		panic(fmt.Errorf("suna.sqls: `%v` is not a pointer", dist))
	}
	dT = dT.Elem()

	var queryFunc func(context.Context, interface{}, string, ...interface{}) error

	switch dT.Kind() {
	case reflect.Slice:
		if dT.Elem().Kind() == reflect.Uint8 { // []byte
			queryFunc = Executor(ctx).GetContext
		} else {
			queryFunc = Executor(ctx).SelectContext
		}
	default:
		queryFunc = Executor(ctx).GetContext
	}

	_DoSqlLogging(query, args)
	err = queryFunc(ctx, dist, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		panic(err)
	}
	return true
}

func ExecuteSql(ctx context.Context, builder ci.Sqlizer) sql.Result {
	query, args, err := builder.ToSql()
	if err != nil {
		panic(err)
	}
	_DoSqlLogging(query, args)
	r, e := Executor(ctx).ExecContext(ctx, query, args...)
	if e != nil {
		panic(e)
	}
	return r
}

func PrepareStmt(ctx context.Context, q string) *sqlx.Stmt {
	_DoSqlLogging("stmt <"+q+">", nil)
	stmt, err := Executor(ctx).PreparexContext(ctx, q)
	if err != nil {
		panic(err)
	}
	return stmt
}
