package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"log"
	"os"
	"reflect"
)

type LoggingExecutor struct {
	Executor
}

var logging = false

func EnableLogging() { logging = true }

var logger *log.Logger

func SetLogger(l *log.Logger) { logger = l }

func init() {
	logger = log.New(os.Stdout, "sha.sqlx ", log.LstdFlags)
}

// scan
func (w LoggingExecutor) ScanRow(ctx context.Context, q string, namedargs interface{}, dist ...interface{}) error {
	q, a := BindNamedArgs(w.Executor, q, namedargs)
	row := w.Executor.QueryRowxContext(ctx, q, a...)
	if err := row.Err(); err != nil {
		return err
	}
	return row.Scan(dist...)
}

func (w LoggingExecutor) ScanRows(ctx context.Context, q string, namedargs interface{}, scanner func(*sqlx.Rows) error) error {
	q, a := BindNamedArgs(w.Executor, q, namedargs)

	rows, err := w.Executor.QueryxContext(ctx, q, a...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = scanner(rows)
		if err != nil {
			return err
		}
	}
	return nil
}

func BindNamedArgs(exe Executor, q string, namedArgs interface{}) (string, []interface{}) {
	var qs string
	var args []interface{}
	var err error
	if namedArgs != nil {
		switch rv := namedArgs.(type) {
		case Data:
			qs, args, err = exe.BindNamed(q, (map[string]interface{})(rv))
		case map[string]interface{}:
			qs, args, err = exe.BindNamed(q, rv)
		default:
			qs, args, err = exe.BindNamed(q, namedArgs)
		}
	} else {
		qs = q
	}
	if err != nil {
		panic(err)
	}
	if logging {
		logger.Printf("%s %v\n", qs, args)
	}
	return qs, args
}

func (w LoggingExecutor) Select(ctx context.Context, q string, namedArgs interface{}, sliceDist interface{}) error {
	q, a := BindNamedArgs(w.Executor, q, namedArgs)
	return w.Executor.SelectContext(ctx, sliceDist, q, a...)
}

func (w LoggingExecutor) Get(ctx context.Context, q string, namedArgs interface{}, dist interface{}) error {
	q, a := BindNamedArgs(w.Executor, q, namedArgs)
	return w.Executor.GetContext(ctx, dist, q, a...)
}

// exec
func (w LoggingExecutor) Exec(ctx context.Context, q string, namedargs interface{}) (sql.Result, error) {
	q, a := BindNamedArgs(w.Executor, q, namedargs)
	return w.Executor.ExecContext(ctx, q, a...)
}

func (w LoggingExecutor) savePoint(ctx context.Context, name string) error {
	_, e := w.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", name), nil)
	return e
}

func (w LoggingExecutor) releaseSavePoint(ctx context.Context, name string) error {
	_, e := w.Exec(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", name), nil)
	return e
}

func (w LoggingExecutor) rollbackToSavePoint(ctx context.Context, name string) error {
	_, e := w.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name), nil)
	return e
}

var mapper = reflectx.NewMapper("db")

func (w LoggingExecutor) JoinGet(
	ctx context.Context, q string, namedArgs interface{}, dists ...interface{},
) error {
	q, a := BindNamedArgs(w.Executor, q, namedArgs)
	row := w.Executor.QueryRowxContext(ctx, q, a...)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return joinScan(row, dists, func(_ interface{}, idx int) interface{} { return dists[idx] })
}

type _RowI interface {
	Scan(...interface{}) error
	Columns() ([]string, error)
}

func joinScan(r _RowI, dist interface{}, get func(interface{}, int) interface{}) error {
	columns, err := r.Columns()
	if err != nil {
		return err
	}

	dIdx := 0
	var d interface{}
	var dV reflect.Value
	var dT *reflectx.StructMap
	var dN int

	var ptrs []interface{}
	for _, c := range columns {
		if dT == nil {
			d = get(dist, dIdx)
			dV = reflect.ValueOf(d).Elem()
			dT = mapper.TypeMap(reflect.TypeOf(d).Elem())
			dN = len(dT.Names)
		}
		fi, ok := dT.Names[c]
		if !ok {
			return fmt.Errorf("sha.sqlx: bad column name `%s`", c)
		}
		f := dV.FieldByIndex(fi.Index)
		ptrs = append(ptrs, f.Addr().Interface())
		dN--
		if dN == 0 {
			dT = nil
			dIdx++
		}
	}
	return r.Scan(ptrs...)
}

func (w LoggingExecutor) JoinSelect(
	ctx context.Context, q string, namedArgs interface{},
	dists interface{}, constructor func(interface{}), get func(interface{}, int) interface{},
) error {
	q, a := BindNamedArgs(w.Executor, q, namedArgs)
	rows, err := w.Executor.QueryxContext(ctx, q, a...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	defer rows.Close()

	distsV := reflect.ValueOf(dists).Elem()
	eleT := reflect.TypeOf(dists).Elem().Elem()
	for rows.Next() {
		if eleT.Kind() == reflect.Ptr {
			ele := reflect.New(eleT.Elem()).Interface()
			if constructor != nil {
				constructor(ele)
			}
			if err := joinScan(rows, ele, get); err != nil {
				return err
			}
			distsV = reflect.Append(distsV, reflect.ValueOf(ele))
		} else {
			l := distsV.Len() + 1
			doAppend := true
			var eleV reflect.Value
			var eleP interface{}
			if l < distsV.Cap() {
				distsV.SetLen(l)
				doAppend = false

				eleV = distsV.Index(l - 1)
				eleP = eleV.Addr().Interface()
			} else {
				elePV := reflect.New(eleT)
				eleP = elePV.Interface()
				eleV = elePV.Elem()
			}

			if constructor != nil {
				constructor(eleP)
			}
			if err := joinScan(rows, eleP, get); err != nil {
				return err
			}
			if doAppend {
				distsV = reflect.Append(distsV, eleV)
			}
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}
	reflect.ValueOf(dists).Elem().Set(distsV)
	return nil
}
