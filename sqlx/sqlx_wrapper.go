package sqlx

import (
	"context"
	"database/sql"
	"log"
)

type W struct {
	Exe Executor
}

var logging = false

func EnableLogging() {
	logging = true
}

// scan
func (w W) Row(ctx context.Context, q string, namedargs interface{}, dist ...interface{}) error {
	q, a := bindNamedargs(w.Exe, q, namedargs)
	row := w.Exe.QueryRowxContext(ctx, q, a...)
	if err := row.Err(); err != nil {
		return err
	}
	return row.Scan(dist...)
}

func bindNamedargs(exe Executor, q string, namedargs interface{}) (string, []interface{}) {
	var qs string
	var args []interface{}
	var err error
	if namedargs != nil {
		switch rv := namedargs.(type) {
		case Data:
			qs, args, err = exe.BindNamed(q, (map[string]interface{})(rv))
		default:
			qs, args, err = exe.BindNamed(q, namedargs)
		}
	} else {
		qs = q
	}
	if err != nil {
		panic(err)
	}
	if logging {
		log.Printf("suna.sqlx: %s %v\n", qs, args)
	}
	return qs, args
}

func (w W) Rows(ctx context.Context, q string, namedargs interface{}, dist interface{}) error {
	q, a := bindNamedargs(w.Exe, q, namedargs)
	return Exe(ctx).Exe.SelectContext(ctx, dist, q, a...)
}

func (w W) RowStruct(ctx context.Context, q string, namedargs interface{}, dist interface{}) error {
	q, a := bindNamedargs(w.Exe, q, namedargs)

	row := w.Exe.QueryRowxContext(ctx, q, a...)
	if err := row.Err(); err != nil {
		return err
	}
	return row.StructScan(dist)
}

func (w W) RowsStruct(ctx context.Context, q string, namedargs interface{}, dist interface{}) error {
	q, a := bindNamedargs(w.Exe, q, namedargs)

	return w.Exe.SelectContext(ctx, dist, q, a...)
}

func (w W) RowsScan(ctx context.Context, q string, namedargs interface{}, scanner Scanner) error {
	q, a := bindNamedargs(w.Exe, q, namedargs)

	rows, err := w.Exe.QueryxContext(ctx, q, a...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return scanner.Scan(rows)
}

// exec
func (w W) Exec(ctx context.Context, q string, namedargs interface{}) sql.Result {
	q, a := bindNamedargs(w.Exe, q, namedargs)
	r, err := w.Exe.ExecContext(ctx, q, a...)
	if err != nil {
		panic(err)
	}
	return r
}
