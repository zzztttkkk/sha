package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"reflect"
	"strings"
)

type Operator struct {
	tablename string
	ele       Modeler
	info      *_StructInfo
}

func NewOperator(ele Modeler) *Operator {
	op := &Operator{ele: ele}
	op.info = getStructInfo(reflect.TypeOf(ele))
	return op
}

func (op *Operator) IsMutableField(f string) bool {
	_, ok := op.info.mutable[f]
	return ok
}

func (op *Operator) IsMutableFieldsData(data Data) bool {
	for k, _ := range data {
		if !op.IsMutableField(k) {
			return false
		}
	}
	return true
}

var ErrImmutableField = errors.New("sha.sqlx: immutable field")
var ErrEmptyConditionOrEmptyData = errors.New("sha.sqlx: empty condition or empty data")

func (op *Operator) MustMutableFieldsData(data Data) {
	if !op.IsMutableFieldsData(data) {
		panic(ErrImmutableField)
	}
}

func (op *Operator) GroupColumns(name string) []string {
	m := op.info.groups[name]
	if len(m) < 1 {
		return nil
	}
	var ret []string
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}

func getTableName(ele Modeler) string { return ele.TableName() }

func (op *Operator) TableName() string {
	if len(op.tablename) < 1 {
		op.tablename = getTableName(op.ele)
	}
	return op.tablename
}

func doCreate(db *sqlx.DB, name string, v Modeler) {
	columns := v.TableColumns(db)
	q := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (%s)",
		name,
		strings.Join(columns, ","),
	)
	if logging {
		logger.Printf("%s\n", q)
	}
	db.MustExec(q)
}

func (op *Operator) CreateTable() { doCreate(writableDb, op.TableName(), op.ele) }

func (op *Operator) simpleSelect(group, cond string) string {
	buf := strings.Builder{}
	buf.WriteString("SELECT ")
	if group == "*" {
		buf.WriteString("*")
	} else {
		keys := op.info.groups[group]
		lastInd := len(keys) - 1
		if lastInd < 0 {
			panic(fmt.Errorf("sha.sqlx: empty key group `%s`", group))
		}
		i := 0
		for v := range keys {
			buf.WriteString(v)
			if i < lastInd {
				buf.WriteRune(',')
			}
			i++
		}
	}
	buf.WriteString(" FROM ")
	buf.WriteString(op.TableName())
	if len(cond) > 0 {
		buf.WriteRune(' ')
		buf.WriteString(cond)
	}
	return buf.String()
}

func (op *Operator) simpleUpdate(cond string, m Data) (string, Data) {
	if len(cond) < 1 || len(m) < 1 {
		panic(ErrEmptyConditionOrEmptyData)
	}

	buf := strings.Builder{}
	buf.WriteString("UPDATE ")
	buf.WriteString(op.TableName())
	buf.WriteString(" SET ")

	var retM Data
	i := 0
	lastInd := len(m) - 1
	for k, v := range m {
		buf.WriteString(k)
		buf.WriteRune('=')

		switch rv := v.(type) {
		case Raw:
			buf.WriteString(string(rv))
		default:
			buf.WriteRune(':')
			buf.WriteString(k)
			if retM == nil {
				retM = Data{}
			}
			retM[k] = v
		}

		if i < lastInd {
			buf.WriteRune(',')
		}
		i++
	}
	buf.WriteString(cond)
	return buf.String(), retM
}

var dataType = reflect.TypeOf(Data{})
var mapType = reflect.TypeOf(map[string]interface{}{})

func mergeMap(ctx context.Context, dist Data, src interface{}) interface{} {
	if len(dist) < 1 {
		return src
	}

	if src == nil {
		return dist
	}

	t := reflect.TypeOf(src)
	if t.Kind() == reflect.Map {
		if t == mapType || t == dataType {
			for a, b := range src.(map[string]interface{}) {
				dist[a] = b
			}
			return dist
		} else {
			panic(errors.New("sha.sqlx: bad data type"))
		}
	}

	t = reflectx.Deref(t)
	if t.Kind() == reflect.Struct {
		val := reflect.ValueOf(src)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		structInfo := db(ctx).Mapper.TypeMap(t)
		for _, f := range structInfo.Index {
			dist[f.Name] = val.FieldByIndex(f.Index).Interface()
		}
	} else {
		panic(errors.New("sha.sqlx: bad data type"))
	}
	return dist
}

func (op *Operator) FetchOne(ctx context.Context, keysGroup string, condition string, namedargs interface{}, dist interface{}) error {
	return Exe(ctx).Get(ctx, op.simpleSelect(keysGroup, condition), namedargs, dist)
}

func (op *Operator) RowColumns(ctx context.Context, columns []string, cond string, namedargs interface{}, dist ...interface{}) error {
	buf := strings.Builder{}
	buf.WriteString("SELECT ")
	buf.WriteString(strings.Join(columns, ", "))
	buf.WriteString(" FROM ")
	buf.WriteString(op.TableName())
	buf.WriteRune(' ')
	if len(cond) > 0 {
		buf.WriteString(cond)
	}
	return Exe(ctx).ScanRow(ctx, buf.String(), namedargs, dist...)
}

func (op *Operator) RowsColumn(ctx context.Context, column string, cond string, namedargs interface{}, dist interface{}) error {
	buf := strings.Builder{}
	buf.WriteString("SELECT ")
	buf.WriteString(column)
	buf.WriteString(" FROM ")
	buf.WriteString(op.TableName())
	buf.WriteRune(' ')
	if len(cond) > 0 {
		buf.WriteString(cond)
	}
	return Exe(ctx).Select(ctx, buf.String(), namedargs, dist)
}

func (op *Operator) FetchMany(ctx context.Context, keysGroup string, condition string, arg interface{}, dist interface{}) error {
	return Exe(ctx).Select(ctx, op.simpleSelect(keysGroup, condition), arg, dist)
}

type Raw string
type Data map[string]interface{}

func (op *Operator) Update(ctx context.Context, data Data, condition string, namedargs interface{}) sql.Result {
	s, m := op.simpleUpdate(condition, data)
	return Exe(ctx).Exec(ctx, s, mergeMap(ctx, m, namedargs))
}

func (op *Operator) simpleInsert(data Data, returning string) (string, Data) {
	if len(data) < 1 {
		panic(ErrEmptyConditionOrEmptyData)
	}

	buf := strings.Builder{}
	buf.WriteString("INSERT INTO ")
	buf.WriteString(op.TableName())
	buf.WriteRune('(')

	var vals []string
	var retM Data
	ind := 0
	lastInd := len(data) - 1
	for a, b := range data {
		switch rv := b.(type) {
		case Raw:
			vals = append(vals, string(rv))
		default:
			vals = append(vals, ":"+a)
			if retM == nil {
				retM = Data{}
			}
			retM[a] = b
		}
		buf.WriteString(a)
		if ind < lastInd {
			buf.WriteRune(',')
		}
		ind++
	}
	buf.WriteString(") VALUES (")
	buf.WriteString(strings.Join(vals, ","))
	buf.WriteRune(')')

	if len(returning) > 0 {
		buf.WriteString(" RETURNING ")
		buf.WriteString(returning)
	}

	return buf.String(), retM
}

func (op *Operator) Insert(ctx context.Context, data Data) int64 {
	s, m := op.simpleInsert(data, "")
	ret, e := Exe(ctx).Exec(ctx, s, m).LastInsertId()
	if e != nil {
		panic(e)
	}
	return ret
}

func (op *Operator) InsertWithReturning(ctx context.Context, data Data, returning string, dist ...interface{}) {
	s, m := op.simpleInsert(data, returning)
	if err := Exe(ctx).ScanRow(ctx, s, m, dist); err != nil {
		panic(err)
	}
}

var ErrDeleteWithoutCondition = errors.New("sha.sqlx: delete without condition")

func (op *Operator) Delete(ctx context.Context, cond string, namedargs interface{}) sql.Result {
	if len(cond) < 1 {
		panic(ErrDeleteWithoutCondition)
	}

	var buf strings.Builder
	buf.WriteString("DELETE FROM ")
	buf.WriteString(op.TableName())
	buf.WriteRune(' ')
	buf.WriteString(cond)

	return Exe(ctx).Exec(ctx, buf.String(), namedargs)
}
