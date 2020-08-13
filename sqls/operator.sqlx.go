package sqls

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/sqrl"
	"github.com/zzztttkkk/suna/sqls/builder"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
)

func (op *Operator) XSelectScan(ctx context.Context, builder *sqrl.SelectBuilder, scanner *Scanner) int {
	q, args, err := builder.ToSql()
	if err != nil {
		panic(err)
	}
	doSqlLog(q, args)
	rows, err := Executor(ctx).QueryxContext(ctx, q, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		panic(err)
	}
	defer rows.Close()

	return scanner.Scan(rows)
}

func (op *Operator) XSelect(ctx context.Context, dist interface{}, builder *sqrl.SelectBuilder) bool {
	q, args, err := builder.ToSql()
	if err != nil {
		panic(err)
	}
	doSqlLog(q, args)

	dT := reflect.TypeOf(dist).Elem()
	switch dT.Kind() {
	case reflect.Slice:
		if dT.Elem().Kind() == reflect.Uint8 { // []byte
			err = Executor(ctx).GetContext(ctx, dist, q, args...)
		} else {
			err = Executor(ctx).SelectContext(ctx, dist, q, args...)
		}
	default:
		err = Executor(ctx).GetContext(ctx, dist, q, args...)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		panic(err)
	}
	return true
}

func (op *Operator) XStmt(ctx context.Context, q string) *sqlx.Stmt {
	doSqlLog("stmt:<"+q+">", nil)
	stmt, err := Executor(ctx).PreparexContext(ctx, q)
	if err != nil {
		panic(err)
	}
	return stmt
}

func (op *Operator) XExecute(ctx context.Context, q string, args ...interface{}) sql.Result {
	doSqlLog(q, args)
	r, e := Executor(ctx).ExecContext(ctx, q, args...)
	if e != nil {
		panic(e)
	}
	return r
}

func (op *Operator) XCreate(ctx context.Context, kvs *utils.Kvs) int64 {
	var ks []string
	var vs []interface{}

	kvs.EachNode(
		func(s string, i interface{}) {
			ks = append(ks, s)
			vs = append(vs, i)
		},
	)

	b := builder.NewInsert(op.tablename).Columns(ks...).Values(vs...)
	if isPostgres {
		b.Returning(op.idField)
	}

	q, args, err := b.ToSql()
	if err != nil {
		panic(err)
	}
	doSqlLog(q, args)

	var lid int64
	var e error
	var r sql.Result

	if isPostgres {
		e = Executor(ctx).GetContext(ctx, &lid, q, args...)
	} else {
		r, e = Executor(ctx).ExecContext(ctx, q, args...)
		if r != nil {
			lid, e = r.LastInsertId()
		}
	}
	if e != nil {
		panic(e)
	}
	return lid
}

func (op *Operator) XUpdate(ctx context.Context, kvs *utils.Kvs, conditions sqrl.Sqlizer, limit int64) int64 {
	b := builder.NewUpdate(op.tablename)
	kvs.EachNode(
		func(s string, i interface{}) {
			b.Set(s, i)
		},
	)
	if conditions != nil {
		b.Where(conditions)
	}
	if limit > -1 {
		b.Limit(uint64(limit))
	}
	q, args, e := b.ToSql()
	if e != nil {
		panic(e)
	}
	doSqlLog(q, args)
	r, e := Executor(ctx).ExecContext(ctx, q, args...)
	if e != nil {
		if e == sql.ErrNoRows {
			return 0
		}
		panic(e)
	}
	c, e := r.RowsAffected()
	if e != nil {
		panic(e)
	}
	return c
}
