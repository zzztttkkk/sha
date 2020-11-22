package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type _ExcScanner struct {
	executor Executor
}

type Sqler interface {
	Sql() (string, []interface{}, error)
}

type StrSql string

func (s StrSql) Sql() (string, []interface{}, error) {
	return string(s), nil, nil
}

func (sc _ExcScanner) Executor() Executor {
	return sc.executor
}

func (sc _ExcScanner) Row(ctx context.Context, sqler Sqler, dist ...interface{}) error {
	s, a, err := sqler.Sql()
	if err != nil {
		panic(err)
	}

	row := sc.executor.QueryRowContext(ctx, s, a...)
	if err = row.Err(); err != nil {
		return err
	}

	if len(dist) > 1 {
		return row.Scan(dist...)
	}

	d := dist[0]
	dt := reflect.TypeOf(d)
	if dt.Kind() != reflect.Ptr {
		panic(fmt.Errorf("suna.sqlx: `%v` is not a pointer", d))
	}

	sv, ok := d.(sql.Scanner)
	if ok {
		return row.Scan(sv)
	}

	st := dt.Elem()
	if st == timeType {
		return row.Scan(d)
	}

	if st.Kind() == reflect.Struct {
		return scanOneStruct(st, d, row)
	}
	return row.Scan(d)
}

type scanner interface {
	Scan(...interface{}) error
}

func scanOneStruct(t reflect.Type, structPtr interface{}, r scanner) error {
	var fieldPtrs []interface{}
	val := reflect.ValueOf(structPtr).Elem()

	for _, field := range GetFields(t) {
		vf := val.FieldByIndex(field.Index)
		fieldPtrs = append(fieldPtrs, vf.Addr().Interface())
	}

	err := r.Scan(fieldPtrs...)
	if err != nil {
		return err
	}
	return nil
}

var isScannerMap sync.Map

func isScanner(t reflect.Type) bool {
	v, ok := isScannerMap.Load(t)
	if ok {
		return v.(bool)
	}

	ele := reflect.New(t).Interface()
	_, ok = ele.(sql.Scanner)
	isScannerMap.Store(t, ok)
	return ok
}

func (sc _ExcScanner) Rows(ctx context.Context, sqler Sqler, dist interface{}) error {
	s, a, err := sqler.Sql()
	if err != nil {
		panic(err)
	}
	rows, err := sc.executor.QueryContext(ctx, s, a...)
	if err != nil {
		return err
	}
	defer rows.Close()

	dt := reflect.TypeOf(dist)
	if dt.Kind() != reflect.Ptr {
		panic(errors.New("suna.sqlx: dist is not a pointer"))
	}
	lt := dt.Elem()
	if lt.Kind() != reflect.Slice {
		panic(errors.New("suna.sqlx: dist is not a slice pointer"))
	}

	et := lt.Elem()
	if et.Kind() == reflect.Uint8 {
		panic(errors.New("suna.sqlx: dist is a uint8 slice pointer"))
	}

	var ret reflect.Type
	if et.Kind() == reflect.Ptr {
		ret = et.Elem()
	}

	isNotScannerStruct := false
	if et.Kind() == reflect.Struct {
		isNotScannerStruct = et != timeType && !isScanner(et)
	} else if ret != nil && ret.Kind() == reflect.Struct {
		isNotScannerStruct = ret != timeType && !isScanner(ret)
	}

	lv := reflect.ValueOf(dist).Elem()

	if !isNotScannerStruct {
		if et.Kind() != reflect.Ptr { // []int
			for rows.Next() {
				ele := reflect.New(et) // *int
				if err = rows.Scan(ele.Interface()); err != nil {
					return err
				}
				lv.Set(reflect.Append(lv, ele.Elem()))
			}
			return rows.Err()
		} else { // []*int
			for rows.Next() {
				ele := reflect.New(ret) // *int
				if err = rows.Scan(ele.Interface()); err != nil {
					return err
				}
				lv.Set(reflect.Append(lv, ele))
			}
			return rows.Err()
		}
	}

	if et.Kind() != reflect.Ptr { // struct
		for rows.Next() {
			ele := reflect.New(et)
			if err = scanOneStruct(et, ele.Interface(), rows); err != nil {
				return err
			}
			lv.Set(reflect.Append(lv, ele.Elem()))
		}
		return rows.Err()
	}

	// struct ptr
	for rows.Next() {
		ele := reflect.New(ret)
		if err = scanOneStruct(ret, ele.Interface(), rows); err != nil {
			return err
		}
		lv.Set(reflect.Append(lv, ele))
	}
	return rows.Err()
}
