package sqlx

import (
	"context"
	"database/sql"
	"log"
)

type W struct {
	exe Executor
}

var logging = false

func EnableLogging() {
	logging = true
}

// scan
func (w W) Row(ctx context.Context, q string, nmap interface{}, dist ...interface{}) error {
	q, a := qa(w.exe, q, nmap)
	row := w.exe.QueryRowxContext(ctx, q, a...)
	if err := row.Err(); err != nil {
		return err
	}
	return row.Scan(dist...)
}

func qa(exe Executor, q string, nmap interface{}) (string, []interface{}) {
	var qs string
	var args []interface{}
	var err error
	if nmap != nil {
		switch rv := nmap.(type) {
		case M:
			qs, args, err = exe.BindNamed(q, (map[string]interface{})(rv))
		default:
			qs, args, err = exe.BindNamed(q, nmap)
		}
	} else {
		qs = q
	}
	if err != nil {
		panic(err)
	}
	if logging {
		log.Printf("suna.sqlx: %s %s\n", qs, args)
	}
	return qs, args
}

func (w W) RowStruct(ctx context.Context, q string, nmap interface{}, dist interface{}) error {
	q, a := qa(w.exe, q, nmap)

	row := w.exe.QueryRowxContext(ctx, q, a...)
	if err := row.Err(); err != nil {
		return err
	}
	return row.StructScan(dist)
}

func (w W) RowsStruct(ctx context.Context, q string, nmap interface{}, dist interface{}) error {
	q, a := qa(w.exe, q, nmap)

	return w.exe.SelectContext(ctx, dist, q, a...)
}

func (w W) Rows(ctx context.Context, q string, nmap interface{}, dist func() []interface{}, after func()) error {
	q, a := qa(w.exe, q, nmap)

	rows, err := w.exe.QueryxContext(ctx, q, a...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(dist()...); err != nil {
			return err
		}
		if after != nil {
			after()
		}
	}
	return nil
}

func (w W) RowsStaticDist(ctx context.Context, q string, nmap interface{}, dist []interface{}, after func()) error {
	q, a := qa(w.exe, q, nmap)

	rows, err := w.exe.QueryxContext(ctx, q, a...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(dist...); err != nil {
			return err
		}
		after()
	}
	return nil
}

// exec
func (w W) Exec(ctx context.Context, q string, nmap interface{}) sql.Result {
	q, a := qa(w.exe, q, nmap)
	r, err := w.exe.ExecContext(ctx, q, a...)
	if err != nil {
		panic(err)
	}
	return r
}
