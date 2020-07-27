package sqls

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zzztttkkk/suna/utils"
	"strings"

	"github.com/jmoiron/sqlx"
)

// X: 	sqlx

// one col one row
func (op *Operator) XQ11(ctx context.Context, dist interface{}, q string, args ...interface{}) bool {
	doSqlLog(q, args)
	if err := Executor(ctx).GetContext(ctx, dist, q, args...); err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		panic(err)
	}
	return true
}

// one col many rows
func (op *Operator) XQ1n(ctx context.Context, silceDist interface{}, q string, args ...interface{}) bool {
	doSqlLog(q, args)
	if err := Executor(ctx).SelectContext(ctx, silceDist, q, args...); err != nil {
		if err == sql.ErrNoRows {
			return true
		}
		panic(err)
	}
	return true
}

func (op *Operator) XQ1nScan(ctx context.Context, dist interface{}, afterScan func() error, q string, args ...interface{}) int64 {
	rows, err := Executor(ctx).QueryxContext(ctx, q, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0
		}
		panic(err)
	}

	var count int64
	for rows.Next() {
		if err = rows.Scan(dist); err != nil {
			panic(err)
		}
		if afterScan != nil {
			if err = afterScan(); err != nil {
				panic(err)
			}
		}
		count++
	}
	return count
}

// many cols one row
func (op *Operator) XQn1(ctx context.Context, dist []interface{}, q string, args ...interface{}) bool {
	doSqlLog(q, args)
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

// many cols many rows, each row use each dists
func (op *Operator) XQnn(ctx context.Context, distsGen func() []interface{}, afterScan func(...interface{}) error, q string, args ...interface{}) int64 {
	doSqlLog(q, args)
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
func (op *Operator) XQnnScan(ctx context.Context, dists []interface{}, afterScan func() error, q string, args ...interface{}) int64 {
	doSqlLog(q, args)
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
func (op *Operator) XQs1(ctx context.Context, dist interface{}, q string, args ...interface{}) bool {
	doSqlLog(q, args)
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
func (op *Operator) XQsn(ctx context.Context, constructor func() interface{}, afterScan func(context.Context, interface{}) error, q string, args ...interface{}) int64 {
	doSqlLog(q, args)
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
func (op *Operator) XQsnScan(ctx context.Context, ele interface{}, afterScan func(context.Context) error, q string, args ...interface{}) int64 {
	doSqlLog(q, args)
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
	op.XQ11(ctx, &c, q, args...)
	return
}

func (op *Operator) XExists(ctx context.Context, key string, condition string, args ...interface{}) bool {
	return op.XCount(ctx, key, condition, args...) > 0
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

func mysqlCreate(ctx context.Context, op *Operator, q string, args []interface{}) int64 {
	result, err := Executor(ctx).ExecContext(ctx, q, args...)
	if err != nil {
		panic(err)
	}
	lid, _ := result.LastInsertId()
	return lid
}

func postgresCreate(ctx context.Context, op *Operator, q string, args []interface{}) int64 {
	q = strings.TrimSpace(q)
	if q[0] != 'i' {
		panic(fmt.Errorf("suna.sqls: not a insert query"))
	}
	if q[len(q)-1] == ';' {
		q = q[:len(q)-1]
	}

	idField := "id"
	if len(op.idField) > 0 {
		idField = op.idField
	}

	q += " returning " + idField
	var lid int64
	err := Executor(ctx).GetContext(ctx, &lid, q, args...)
	if err != nil {
		panic(err)
	}
	return lid
}

func (op *Operator) XCreate(ctx context.Context, m utils.M) int64 {
	var ks []string
	var pls []string

	for k, v := range m {
		ks = append(ks, k)
		switch rv := v.(type) {
		case Raw:
			pls = append(pls, string(rv))
		default:
			pls = append(pls, ":"+k)
		}
	}

	s, vl := op.BindNamed(
		fmt.Sprintf(
			"insert into %s (%s) values (%s)",
			op.tablename, strings.Join(ks, ","),
			strings.Join(pls, ","),
		),
		m,
	)

	return doCreate(ctx, op, s, vl)
}

func (op *Operator) XUpdate(ctx context.Context, updates utils.M, condition string, conditionArgs ...interface{}) int64 {
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
	vl = append(vl, conditionArgs...)

	if sqlLog {
		doSqlLog(q, conditionArgs)
	}

	result, err := Executor(ctx).ExecContext(ctx, q, append(vl, conditionArgs)...)
	if err != nil {
		panic(err)
	}

	count, _ := result.RowsAffected()
	return count
}
