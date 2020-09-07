package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/sqlx/sqlr"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
)

func (op *Operator) ExecuteScan(ctx context.Context, builder *sqlr.SelectBuilder, scanner *Scanner) int {
	if len(builder.Tables()) < 1 {
		builder.From(op.TableName())
	}
	q, args, err := builder.ToSql()
	if err != nil {
		panic(err)
	}
	_DoSqlLogging(q, args)
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

func (op *Operator) ExecuteSelect(ctx context.Context, dist interface{}, builder *sqlr.SelectBuilder) bool {
	if len(builder.Tables()) < 1 {
		builder.From(op.TableName())
	}

	dT := reflect.TypeOf(dist)
	if dT.Kind() != reflect.Ptr {
		panic(fmt.Errorf("suna.sqls: `%v` is not a pointer", dist))
	}
	var queryFunc func(context.Context, interface{}, string, ...interface{}) error
	switch dT.Elem().Kind() {
	case reflect.Slice:
		if dT.Elem().Kind() == reflect.Uint8 { // []byte
			queryFunc = Executor(ctx).GetContext
			builder.Limit(1)
		} else {
			queryFunc = Executor(ctx).SelectContext
		}
	default:
		queryFunc = Executor(ctx).GetContext
		builder.Limit(1)
	}

	q, args, err := builder.ToSql()
	if err != nil {
		panic(err)
	}
	_DoSqlLogging(q, args)

	err = queryFunc(ctx, dist, q, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		panic(err)
	}
	return true
}

func (op *Operator) ExecuteExistsTest(ctx context.Context, conditions sqlr.Sqlizer) bool {
	c := 0
	op.ExecuteSelect(ctx, &c, op.SelectBuilder(ctx, "count(*)").From(op.TableName()).Where(conditions))
	return c > 0
}

func (op *Operator) PrepareStmt(ctx context.Context, q string) *sqlx.Stmt {
	_DoSqlLogging("stmt <"+q+">", nil)
	stmt, err := Executor(ctx).PreparexContext(ctx, q)
	if err != nil {
		panic(err)
	}
	return stmt
}

func (op *Operator) ExecuteSql(ctx context.Context, q string, args ...interface{}) sql.Result {
	_DoSqlLogging(q, args)
	r, e := Executor(ctx).ExecContext(ctx, q, args...)
	if e != nil {
		panic(e)
	}
	return r
}

func (op *Operator) ExecuteCreate(ctx context.Context, kvs *utils.Kvs) int64 {
	var ks []string
	var vs []interface{}

	kvs.EachNode(
		func(s string, i interface{}) {
			ks = append(ks, s)
			vs = append(vs, i)
		},
	)

	E := Executor(ctx)
	isPostgres := E.DriverName() == "postgres"

	b := op.InsertBuilder(ctx, op.TableName()).Columns(ks...).Values(vs...)
	if isPostgres {
		if len(op.idField) > 0 {
			b.Returning(op.idField)
		} else {
			b.Returning("id")
		}
	}

	q, args, err := b.ToSql()
	if err != nil {
		panic(err)
	}
	_DoSqlLogging(q, args)

	var lid int64
	var e error
	var r sql.Result

	if isPostgres {
		e = E.GetContext(ctx, &lid, q, args...)
	} else {
		r, e = E.ExecContext(ctx, q, args...)
		if r != nil {
			lid, e = r.LastInsertId()
		}
	}
	if e != nil {
		panic(e)
	}
	return lid
}

type Where struct {
	s     string
	args  []interface{}
	limit int64
}

func (w *Where) Limit(limit int64) *Where {
	w.limit = limit
	return w
}

func NewWhere(s string, args ...interface{}) *Where { return &Where{s: s, args: args} }

func (op *Operator) ExecuteUpdate(ctx context.Context, kvs *utils.Kvs, where *Where) int64 {
	update := op.UpdateBuilder(ctx, op.tablename)
	kvs.EachNode(
		func(s string, i interface{}) {
			update.Set(s, i)
		},
	)
	if where == nil || len(where.s) < 1 {
		panic("suna.sqls: update without any conditions")
	}
	update.Where(where.s, where.args...)
	if where.limit > 0 {
		update.Limit(uint64(where.limit))
	}
	q, args, e := update.ToSql()
	if e != nil {
		panic(e)
	}

	r := op.ExecuteSql(ctx, q, args...)
	v, e := r.RowsAffected()
	if e != nil {
		panic(e)
	}
	return v
}

func (op *Operator) ExecuteUpdate1(ctx context.Context, kvs *utils.Kvs, condition *_Conditions, limit int64) int64 {
	update := op.UpdateBuilder(ctx, op.tablename)
	kvs.EachNode(
		func(s string, i interface{}) {
			update.Set(s, i)
		},
	)
	if condition == nil {
		panic("suna.sqls: update without any conditions")
	}
	update.Where(condition)
	if limit > 0 {
		update.Limit(uint64(limit))
	}
	q, args, e := update.ToSql()
	if e != nil {
		panic(e)
	}
	r := op.ExecuteSql(ctx, q, args...)
	v, e := r.RowsAffected()
	if e != nil {
		panic(e)
	}
	return v
}
