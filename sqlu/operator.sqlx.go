package sqlu

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zzztttkkk/suna/utils"
	"strings"

	"github.com/jmoiron/sqlx"
)

// X: 	sqlx
// L: 	literal
// S: 	struct
// Ls: 	literals
// O:  	one row
// M:  	many rows

// one col single row
func (op *Operator) XLO(ctx context.Context, dist interface{}, q string, args ...interface{}) bool {
	if sqlLog {
		doSqlLog(q, args)
	}

	if err := Executor(ctx).GetContext(ctx, dist, q, args...); err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		panic(err)
	}
	return true
}

// one col multi rows
func (op *Operator) XLM(ctx context.Context, silceDist interface{}, q string, args ...interface{}) bool {
	if sqlLog {
		doSqlLog(q, args)
	}
	if err := Executor(ctx).SelectContext(ctx, silceDist, q, args...); err != nil {
		if err == sql.ErrNoRows {
			return true
		}
		panic(err)
	}
	return true
}

// many cols single row
func (op *Operator) XLsO(ctx context.Context, dist []interface{}, q string, args ...interface{}) bool {
	if sqlLog {
		doSqlLog(q, args)
	}
	row := Executor(ctx).QueryRowxContext(ctx, q, args...)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		panic(err)
	}
	if err := row.Scan(dist...); err != nil {
		panic(err)
	}
	return true
}

// many cols multi rows, each row use each dists
func (op *Operator) XLsM(ctx context.Context, distsGen func() []interface{}, afterScan func(...interface{}) error, q string, args ...interface{}) int64 {
	if sqlLog {
		doSqlLog(q, args)
	}
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
		vl := distsGen()
		if err := rows.Scan(vl...); err != nil {
			panic(err)
		}
		if afterScan != nil {
			if err := afterScan(vl...); err != nil {
				panic(err)
			}
		}
		count++
	}
	return count
}

// many cols multi rows, each row use same dists
func (op *Operator) XLsMSimple(ctx context.Context, dists []interface{}, afterScan func() error, q string, args ...interface{}) int64 {
	if sqlLog {
		doSqlLog(q, args)
	}
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
		if err := rows.Scan(dists...); err != nil {
			panic(err)
		}
		if afterScan != nil {
			if err := afterScan(); err != nil {
				panic(err)
			}
		}
		count++
	}
	return count
}

// struct single row
func (op *Operator) XSO(ctx context.Context, dist interface{}, q string, args ...interface{}) bool {
	if sqlLog {
		doSqlLog(q, args)
	}
	row := Executor(ctx).QueryRowxContext(ctx, q, args...)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		panic(err)
	}
	if err := row.StructScan(dist); err != nil {
		panic(err)
	}
	return true
}

// struct multi rows, each row use each dist
func (op *Operator) XSM(ctx context.Context, constructor func() interface{}, afterScan func(context.Context, interface{}) error, q string, args ...interface{}) int64 {
	if sqlLog {
		doSqlLog(q, args)
	}
	rows, err := Executor(ctx).QueryxContext(ctx, q, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		panic(err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		ele := constructor()
		err := rows.StructScan(ele)
		if err != nil {
			panic(err)
		}
		if afterScan != nil {
			if err = afterScan(ctx, ele); err != nil {
				panic(err)
			}
		}
		count++
	}
	return count
}

// struct multi row, each row use same dist
func (op *Operator) XSMSimple(ctx context.Context, ele interface{}, afterScan func(context.Context) error, q string, args ...interface{}) int64 {
	if sqlLog {
		doSqlLog(q, args)
	}
	rows, err := Executor(ctx).QueryxContext(ctx, q, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		panic(err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		err := rows.StructScan(ele)
		if err != nil {
			panic(err)
		}
		if afterScan != nil {
			if err = afterScan(ctx); err != nil {
				panic(err)
			}
		}
		count++
	}
	return count
}

func (op *Operator) XCount(ctx context.Context, key string, condition string, args ...interface{}) (c int64) {
	q := fmt.Sprintf(`select count(%s) from %s`, key, op.tablename)
	if len(condition) > 0 {
		q += " where " + condition
	}
	op.XLO(ctx, &c, q, args...)
	return
}

func (op *Operator) XExists(ctx context.Context, key string, condition string, args ...interface{}) bool {
	return op.XCount(ctx, key, condition, args...) > 0
}

func (op *Operator) XStmt(ctx context.Context, q string) *sqlx.Stmt {
	if sqlLog {
		doSqlLog("stmt:<"+q+">", nil)
	}

	stmt, err := Executor(ctx).PreparexContext(ctx, q)
	if err != nil {
		panic(err)
	}
	return stmt
}

func (op *Operator) XExecute(ctx context.Context, q string, args ...interface{}) sql.Result {
	if sqlLog {
		doSqlLog(q, args)
	}
	r, e := Executor(ctx).ExecContext(ctx, q, args...)
	if e != nil {
		panic(e)
	}
	return r
}

func (op *Operator) XCreate(ctx context.Context, m utils.M) int64 {
	var ks []string
	var pls []string
	var vl []interface{}

	for k, v := range m {
		ks = append(ks, k)

		switch rv := v.(type) {
		case Raw:
			pls = append(pls, string(rv))
		default:
			pls = append(pls, ":"+k)
			vl = append(vl, v)
		}
	}

	s := fmt.Sprintf(
		"insert into %s (%s) values (%s)",
		op.tablename, strings.Join(ks, ","),
		strings.Join(pls, ","),
	)

	r := op.XExecute(ctx, s, m)
	lid, err := r.LastInsertId()
	if err != nil {
		panic(err)
	}
	return lid
}

func (op *Operator) XUpdate(ctx context.Context, updates utils.M, condition string, args ...interface{}) int64 {
	var pls []string
	var vl []interface{}

	for k, v := range updates {
		switch rv := v.(type) {
		case Raw:
			pls = append(pls, fmt.Sprintf("%s=%s", k, string(rv)))
		default:
			pls = append(pls, fmt.Sprintf("%s=:%s", k, k))
			vl = append(vl, v)
		}
	}

	s := fmt.Sprintf("update %s set %s", op.TableName(), strings.Join(pls, ","))
	if len(condition) > 0 {
		s += " where " + condition
	}

	q, vl := op.BindNamed(s, updates)
	vl = append(vl, args...)

	r := op.XExecute(ctx, q, vl...)
	count, err := r.RowsAffected()
	if err != nil {
		panic(err)
	}
	return count
}
