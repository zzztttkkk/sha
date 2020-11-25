package sqlx

import (
	"context"
	"database/sql"
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

func (op *Operator) GroupKeys(name string) []string {
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

func toTablename(v string) string {
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
	return toTablename(ele.Type().Name())
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
	buf.WriteRune(' ')
	buf.WriteString(cond)
	return buf.String()
}

func (op *Operator) simpleUpdate(cond string, m M) (string, M) {
	buf := strings.Builder{}
	buf.WriteString("UPDATE ")
	buf.WriteString(op.TableName())
	buf.WriteString(" SET ")

	var retM M
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
				retM = M{}
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

var mType = reflect.TypeOf(M{})

func mergeMap(ctx context.Context, dist M, src interface{}) interface{} {
	if len(dist) < 1 {
		return src
	}

	t := reflect.TypeOf(src)
	if t.Kind() == reflect.Map {
		if t == mType {
			for a, b := range src.(M) {
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

func (op *Operator) One(ctx context.Context, group string, condition string, arg interface{}, dist interface{}) error {
	return Exe(ctx).RowStruct(ctx, op.simpleSelect(group, condition), arg, dist)
}

func (op *Operator) Many(ctx context.Context, group string, condition string, arg interface{}, dist interface{}) error {
	return Exe(ctx).RowsStruct(ctx, op.simpleSelect(group, condition), arg, dist)
}

type Raw string
type M map[string]interface{}

func (op *Operator) Update(ctx context.Context, data M, condition string, arg interface{}) sql.Result {
	s, m := op.simpleUpdate(condition, data)
	return Exe(ctx).Exec(ctx, s, mergeMap(ctx, m, arg))
}

func (op *Operator) simpleInsert(data M, returning string) (string, M) {
	buf := strings.Builder{}
	buf.WriteString("INSERT INTO ")
	buf.WriteString(op.TableName())
	buf.WriteRune('(')

	var vals []string
	var retM M
	ind := 0
	lastInd := len(data) - 1
	for a, b := range data {
		switch rv := b.(type) {
		case Raw:
			vals = append(vals, string(rv))
		default:
			vals = append(vals, ":"+a)
			if retM == nil {
				retM = M{}
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

func (op *Operator) Insert(ctx context.Context, data M) sql.Result {
	s, m := op.simpleInsert(data, "")
	return Exe(ctx).Exec(ctx, s, m)
}

func (op *Operator) InsertWithReturning(ctx context.Context, data M, returning string, dist interface{}) {
	s, m := op.simpleInsert(data, returning)
	if err := Exe(ctx).Row(ctx, s, m, dist); err != nil {
		panic(err)
	}
}
