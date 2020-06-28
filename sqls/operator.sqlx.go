package sqls

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

func (op *Operator) XFetchOne(ctx context.Context, dist interface{}, q string, args ...interface{}) bool {
	if err := Executor(ctx).GetContext(ctx, dist, q, args...); err != nil {
		if err != sql.ErrNoRows {
			return false
		}
		panic(err)
	}
	return true
}

func (op *Operator) XExists(ctx context.Context, q string, args ...interface{}) bool {
	var x int
	op.XFetchOne(ctx, &x, q, args...)
	return x > 0
}

func (op *Operator) XScanOne(ctx context.Context, dist []interface{}, q string, args ...interface{}) bool {
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

func (op *Operator) XScanStructOne(ctx context.Context, dist interface{}, q string, args ...interface{}) bool {
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

func (op *Operator) XScanMany(ctx context.Context, fn func() []interface{}, q string, args ...interface{}) int64 {
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

func (op *Operator) XStructScanMany(ctx context.Context, fn func() interface{}, q string, args ...interface{}) int64 {
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

func (op *Operator) XFetchMany(ctx context.Context, slicePtr interface{}, q string, args ...interface{}) {
	if err := Executor(ctx).SelectContext(ctx, slicePtr, q, args...); err != nil {
		panic(err)
	}
}

func (op *Operator) XStmt(ctx context.Context, q string) *sqlx.Stmt {
	stmt, err := Executor(ctx).PreparexContext(ctx, q)
	if err != nil {
		panic(err)
	}
	return stmt
}

func (op *Operator) XExecute(ctx context.Context, q string, args ...interface{}) sql.Result {
	r, e := Executor(ctx).ExecContext(ctx, q, args...)
	if e != nil {
		panic(e)
	}
	return r
}

func (op *Operator) XCreate(ctx context.Context, dict Dict) int64 {
	query, values := dict.ForCreate(op.ddl.tableName)
	r := op.XExecute(ctx, query, values...)
	lid, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}
	return lid
}

func (op *Operator) XUpdate(ctx context.Context, q string, args ...interface{}) int64 {
	r := op.XExecute(ctx, q, args...)
	rows, err := r.RowsAffected()
	if err != nil {
		panic(err)
	}
	return rows
}
