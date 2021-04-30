package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
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
