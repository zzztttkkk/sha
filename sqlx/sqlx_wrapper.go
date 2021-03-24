package sqlx

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
)

type W struct {
	Raw Executor
}

var logging = false

func EnableLogging() {
	logging = true
}

var logger *log.Logger

func SetLogger(l *log.Logger) {
	logger = l
}

func init() {
	logger = log.New(os.Stdout, "sha.sqlx ", log.LstdFlags)
}

// scan
func (w W) ScanRow(ctx context.Context, q string, namedargs interface{}, dist ...interface{}) error {
	q, a := BindNamedArgs(w.Raw, q, namedargs)
	row := w.Raw.QueryRowxContext(ctx, q, a...)
	if err := row.Err(); err != nil {
		return err
	}
	return row.Scan(dist...)
}

func (w W) ScanRows(ctx context.Context, q string, namedargs interface{}, scanner func(*sqlx.Rows) error) error {
	q, a := BindNamedArgs(w.Raw, q, namedargs)

	rows, err := w.Raw.QueryxContext(ctx, q, a...)
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

func (w W) Select(ctx context.Context, q string, namedArgs interface{}, sliceDist interface{}) error {
	q, a := BindNamedArgs(w.Raw, q, namedArgs)
	return Exe(ctx).Raw.SelectContext(ctx, sliceDist, q, a...)
}

func (w W) Get(ctx context.Context, q string, namedArgs interface{}, dist interface{}) error {
	q, a := BindNamedArgs(w.Raw, q, namedArgs)
	return Exe(ctx).Raw.GetContext(ctx, dist, q, a...)
}

// exec
func (w W) Exec(ctx context.Context, q string, namedargs interface{}) sql.Result {
	q, a := BindNamedArgs(w.Raw, q, namedargs)
	r, err := w.Raw.ExecContext(ctx, q, a...)
	if err != nil {
		panic(err)
	}
	return r
}
