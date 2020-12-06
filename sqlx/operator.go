package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"log"
	"reflect"
	"strings"
	"unicode"
)

type Operator struct {
	tablename string
	ele       reflect.Value
	info      *_StructInfo
}

func NewOperator(ele interface{}) *Operator {
	op := &Operator{ele: reflect.ValueOf(ele)}
	op.info = getStructInfo(reflect.TypeOf(ele))
	return op
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

func toTableName(v string) string {
	var ret []rune
	for i, r := range v {
		if i == 0 {
			ret = append(ret, unicode.ToLower(r))
			continue
		}

		if unicode.IsUpper(r) {
			ret = append(ret, '_', unicode.ToLower(r))
		} else {
			ret = append(ret, r)
		}
	}
	return string(ret)
}

func getTableName(ele reflect.Value) string {
	tablenameFn := ele.MethodByName("TableName")
	if tablenameFn.IsValid() {
		return (tablenameFn.Call(nil)[0]).Interface().(string)
	}
	return toTableName(ele.Type().Name())
}

func (op *Operator) TableName() string {
	if len(op.tablename) < 1 {
		op.tablename = getTableName(op.ele)
	}
	return op.tablename
}

func doCreate(db *sqlx.DB, name string, fnV reflect.Value) {
	columns := fnV.Call([]reflect.Value{reflect.ValueOf(db)})[0].Interface().([]string)
	q := fmt.Sprintf(
		"create table if not exists %s (%s)",
		name,
		strings.Join(columns, ","),
	)
	if logging {
		log.Printf("suna.sqlx: %s\n", q)
	}
	db.MustExec(q)
}

func (op *Operator) CreateTable(forEveryDB bool) {
	fnV := op.ele.MethodByName("TableColumns")
	if !fnV.IsValid() {
		panic(
			fmt.Errorf(
				"suna.sqlx: `%s.%s` does not have a method named `TableColumns`",
				op.ele.Type().PkgPath(),
				op.ele.Type().Name(),
			),
		)
	}

	doCreate(wdb, op.TableName(), fnV)
	if forEveryDB {
		for _, d := range rdbs {
			doCreate(d, op.TableName(), fnV)
		}
	}
}

func (op *Operator) simpleSelect(group, cond string) string {
	buf := strings.Builder{}
	buf.WriteString("SELECT ")
	if group == "*" {
		buf.WriteString("*")
	} else {
		keys := op.info.groups[group]
		lastInd := len(keys) - 1
		if lastInd < 0 {
			panic(fmt.Errorf("suna.sqlx: empty group `%s`", group))
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

var mType = reflect.TypeOf(Data{})

func mergeMap(ctx context.Context, dist Data, src interface{}) interface{} {
	if len(dist) < 1 {
		return src
	}

	t := reflect.TypeOf(src)
	if t.Kind() == reflect.Map {
		if t == mType {
			for a, b := range src.(Data) {
				dist[a] = b
			}
			return dist
		}
		for a, b := range src.(map[string]interface{}) {
			dist[a] = b
		}
		return dist
	}

	t = reflectx.Deref(t)
	if t.Kind() == reflect.Struct {
		val := reflect.ValueOf(src)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		structinfo := db(ctx).Mapper.TypeMap(t)
		for _, f := range structinfo.Index {
			dist[f.Name] = val.FieldByIndex(f.Index).Interface()
		}
	}
	return dist
}

func (op *Operator) FetchOne(ctx context.Context, keysGroup string, condition string, namedargs interface{}, dist interface{}) error {
	return Exe(ctx).RowStruct(ctx, op.simpleSelect(keysGroup, condition), namedargs, dist)
}

func (op *Operator) RowColumns(ctx context.Context, columns string, cond string, namedargs interface{}, dist ...interface{}) error {
	buf := strings.Builder{}
	buf.WriteString("SELECT ")
	buf.WriteString(columns)
	buf.WriteString(" FROM ")
	buf.WriteString(op.TableName())
	buf.WriteRune(' ')
	if len(cond) > 0 {
		buf.WriteString(cond)
	}
	return Exe(ctx).Row(ctx, buf.String(), namedargs, dist...)
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
	return Exe(ctx).Rows(ctx, buf.String(), namedargs, dist)
}

func (op *Operator) FetchMany(ctx context.Context, keysGroup string, condition string, arg interface{}, dist interface{}) error {
	return Exe(ctx).RowsStruct(ctx, op.simpleSelect(keysGroup, condition), arg, dist)
}

type Raw string
type Data map[string]interface{}

func (op *Operator) Update(ctx context.Context, data Data, condition string, namedargs interface{}) sql.Result {
	s, m := op.simpleUpdate(condition, data)
	return Exe(ctx).Exec(ctx, s, mergeMap(ctx, m, namedargs))
}

func (op *Operator) simpleInsert(data Data, returning string) (string, Data) {
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
	if err := Exe(ctx).Row(ctx, s, m, dist); err != nil {
		panic(err)
	}
}

var ErrDeleteWithoutCondition = errors.New("suna.sqlx: delete without condition")

func (op *Operator) Delete(ctx context.Context, cond string, namedargs interface{}) sql.Result {
	if len(cond) < 1 {
		panic(ErrDeleteWithoutCondition)
	}

	var buf strings.Builder
	buf.WriteString("delete from ")
	buf.WriteString(op.TableName())
	buf.WriteRune(' ')

	return Exe(ctx).Exec(ctx, buf.String(), namedargs)
}
