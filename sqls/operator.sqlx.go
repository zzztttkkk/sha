package sqls

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

func (op *Operator) SqlxFetchOne(ctx context.Context, item interface{}, q string, args ...interface{}) bool {
	if err := Executor(ctx).GetContext(ctx, item, q, args...); err != nil {
		if err != sql.ErrNoRows {
			return false
		}
		panic(err)
	}
	return true
}

func (op *Operator) SqlxExists(ctx context.Context, q string, args ...interface{}) bool {
	var x int
	op.SqlxFetchOne(ctx, &x, q, args...)
	return x > 0
}

func (op *Operator) SqlxScanRow(ctx context.Context, dist []interface{}, q string, args ...interface{}) bool {
	row := Executor(ctx).QueryRowxContext(ctx, q, args...)
	if err := row.Err(); err != nil {
		if err != sql.ErrNoRows {
			return false
		}
		panic(err)
	}
	if err := row.Scan(dist...); err != nil {
		panic(err)
	}
	return true
}

func (op *Operator) SqlxScanStructRow(ctx context.Context, dist interface{}, q string, args ...interface{}) bool {
	row := Executor(ctx).QueryRowxContext(ctx, q, args...)
	if err := row.Err(); err != nil {
		if err != sql.ErrNoRows {
			return false
		}
		panic(err)
	}
	if err := row.StructScan(dist); err != nil {
		panic(err)
	}
	return true
}

func (op *Operator) SqlxScanRows(ctx context.Context, fn func() []interface{}, q string, args ...interface{}) int64 {
	var count int64

	rows, err := Executor(ctx).QueryxContext(ctx, q, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return count
		}
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(fn()...)
		count++
		if err != nil {
			panic(err)
		}
	}
	return count
}

func (op *Operator) SqlxStructScanRows(ctx context.Context, fn func() interface{}, q string, args ...interface{}) int64 {
	var count int64

	rows, err := Executor(ctx).QueryxContext(ctx, q, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return count
		}
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.StructScan(fn())
		count++
		if err != nil {
			panic(err)
		}
	}
	return count
}

func (op *Operator) SqlxFetch(ctx context.Context, item interface{}, q string, args ...interface{}) {
	if err := Executor(ctx).SelectContext(ctx, item, q, args...); err != nil {
		panic(err)
	}
}

func (op *Operator) SqlxStmt(ctx context.Context, q string) *sqlx.Stmt {
	stmt, err := Executor(ctx).PreparexContext(ctx, q)
	if err != nil {
		panic(err)
	}
	return stmt
}

func (op *Operator) SqlxExecute(ctx context.Context, q string, args ...interface{}) sql.Result {
	r, e := Executor(ctx).ExecContext(ctx, q, args...)
	if e != nil {
		panic(e)
	}
	return r
}

func (op *Operator) SqlxCreate(ctx context.Context, q string, args ...interface{}) int64 {
	r := op.SqlxExecute(ctx, q, args...)
	lid, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}
	return lid
}

var emptySqlsDictError = errors.New("snow.sqls: empty dict")

func (op *Operator) SqlxUpdate(ctx context.Context, q string, dict Dict, args ...interface{}) int64 {
	for _, k := range op.ddl.consts {
		delete(dict, k)
	}
	if len(dict) < 1 {
		panic(emptySqlsDictError)
	}

	placeholder, values := dict.ForUpdate()
	values = append(values, args...)
	r := op.SqlxExecute(ctx, fmt.Sprintf(q, placeholder), values...)
	rows, err := r.RowsAffected()
	if err != nil {
		panic(err)
	}
	return rows
}
