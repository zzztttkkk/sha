package sqls

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

func (op *Operator) SqlxFetchOne(ctx context.Context, dist interface{}, q string, args ...interface{}) bool {
	if err := Executor(ctx).GetContext(ctx, dist, q, args...); err != nil {
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

func (op *Operator) SqlxScanOne(ctx context.Context, dist []interface{}, q string, args ...interface{}) bool {
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

func (op *Operator) SqlxScanStructOne(ctx context.Context, dist interface{}, q string, args ...interface{}) bool {
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

func (op *Operator) SqlxScanMany(ctx context.Context, fn func() []interface{}, q string, args ...interface{}) int64 {
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

func (op *Operator) SqlxStructScanMany(ctx context.Context, fn func() interface{}, q string, args ...interface{}) int64 {
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

func (op *Operator) SqlxFetchMany(ctx context.Context, slicePtr interface{}, q string, args ...interface{}) {
	if err := Executor(ctx).SelectContext(ctx, slicePtr, q, args...); err != nil {
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

func (op *Operator) SqlxCreate(ctx context.Context, dict Dict) int64 {
	query, values := dict.ForCreate(op.ddl.tableName)
	r := op.SqlxExecute(ctx, query, values...)
	lid, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}
	return lid
}

func (op *Operator) SqlxUpdate(ctx context.Context, q string, args ...interface{}) int64 {
	r := op.SqlxExecute(ctx, q, args...)
	rows, err := r.RowsAffected()
	if err != nil {
		panic(err)
	}
	return rows
}
