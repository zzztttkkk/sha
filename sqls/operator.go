package sqls

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

type Operator struct{}

func (op *Operator) FetchOne(ctx context.Context, item interface{}, q string, args ...interface{}) {
	if err := Executor(ctx).GetContext(ctx, item, q, args...); err != nil {
		panic(err)
	}
}

func (op *Operator) Fetch(ctx context.Context, item interface{}, q string, args ...interface{}) {
	if err := Executor(ctx).SelectContext(ctx, item, q, args...); err != nil {
		panic(err)
	}
}

func (op *Operator) Stmt(ctx context.Context, q string) *sqlx.Stmt {
	stmt, err := Executor(ctx).PreparexContext(ctx, q)
	if err != nil {
		panic(err)
	}
	return stmt
}

func (op *Operator) Execute(ctx context.Context, q string, args ...interface{}) sql.Result {
	r, e := Executor(ctx).ExecContext(ctx, q, args...)
	if e != nil {
		panic(e)
	}
	return r
}

func (op *Operator) Create(ctx context.Context, q string, args ...interface{}) int64 {
	r := op.Execute(ctx, q, args...)
	lid, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}
	return lid
}

func (op *Operator) Update(ctx context.Context, q string, args ...interface{}) int64 {
	r := op.Execute(ctx, q, args...)
	rows, err := r.RowsAffected()
	if err != nil {
		panic(err)
	}
	return rows
}
