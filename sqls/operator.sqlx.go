package sqls

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

func (op *Operator) XStructScanMany(ctx context.Context, constructor func() interface{}, q string, args ...interface{}) int64 {
	return op.XStructScanManyWithInit(ctx, constructor, nil, q, args...)
}

func (op *Operator) XStructScanManyWithInit(ctx context.Context, constructor func() interface{}, initer func(context.Context, interface{}) error, q string, args ...interface{}) int64 {
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
		ele := constructor()

		err := rows.StructScan(ele)
		if err != nil {
			panic(err)
		}

		if initer != nil {
			if err = initer(ctx, ele); err != nil {
				panic(err)
			}
		}

		count++
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

func (op *Operator) XExecute(ctx context.Context, q string, dict Dict) sql.Result {
	s, vl := BindNamed(q, dict)
	r, e := Executor(ctx).ExecContext(ctx, s, vl...)
	if e != nil {
		panic(e)
	}
	return r
}

func (op *Operator) XCreate(ctx context.Context, dict Dict) int64 {
	var ks []string
	var pls []string
	var vl []interface{}

	for k, v := range dict {
		ks = append(ks, k)
		pls = append(pls, ":"+k)
		vl = append(vl, v)
	}

	s := fmt.Sprintf(
		"insert into %s (%s) values (%s)",
		op.tablename, strings.Join(ks, ","),
		strings.Join(pls, ","),
	)

	r := op.XExecute(ctx, s, dict)
	lid, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}
	return lid
}

func (op *Operator) XUpdate(ctx context.Context, updates Dict, condition string, conditionV Dict) int64 {
	var pls []string
	var vl []interface{}

	for k, v := range updates {
		switch rv := v.(type) {
		case Raw:
			pls = append(pls, k+"="+string(rv))
		default:
			pls = append(pls, k+"= :"+k)
			vl = append(vl, v)
		}
	}

	s := fmt.Sprintf("update %s set %s", op.TableName(), strings.Join(pls, ","))
	if len(condition) > 0 {
		s += " where " + condition
		for k, v := range conditionV {
			_, ok := updates[k]
			if ok {
				panic(fmt.Errorf("suna.sqls: conflict key `%s`", k))
			}
			updates[k] = v
		}
	}

	r := op.XExecute(ctx, s, updates)
	count, err := r.RowsAffected()
	if err != nil {
		panic(err)
	}
	return count
}
