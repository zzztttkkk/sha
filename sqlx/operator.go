package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/zzztttkkk/sha/utils"
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

func (op *Operator) IsImmutableField(f string) bool { return op.info.immutable.has(f) }

var ErrImmutableField = errors.New("sha.sqlx: immutable field")
var ErrEmptyConditionOrEmptyData = errors.New("sha.sqlx: empty condition or empty data")

func (op *Operator) GroupColumns(name string) []string {
	s := op.info.groups[name]
	if s == nil {
		return nil
	}
	return s.all()
}

func (op *Operator) GroupColumnsAppend(name, val string) {
	s := op.info.groups[name]
	if s == nil {
		s = &_StringSet{}
		op.info.groups[name] = s
	}
	s.add(val)
}

func (op *Operator) GroupColumnsRemove(name, val string) {
	s := op.info.groups[name]
	if s == nil {
		return
	}
	s.del(val)
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

func (op *Operator) simpleSelect(driver, group, cond string) string {
	buf := strings.Builder{}
	buf.WriteString("SELECT ")
	if group == "*" {
		buf.WriteString("*")
	} else {
		keys := op.info.groups[group].all()
		lastInd := len(keys) - 1
		if lastInd < 0 {
			panic(fmt.Errorf("sha.sqlx: empty key group `%s`", group))
		}
		i := 0
		for _, k := range keys {
			writeIdentifier(driver, k, &buf)
			if i < lastInd {
				buf.WriteRune(',')
			}
			i++
		}
	}
	buf.WriteString(" FROM ")
	writeIdentifier(driver, op.TableName(), &buf)
	if len(cond) > 0 {
		buf.WriteRune(' ')
		buf.WriteString(cond)
	}
	return buf.String()
}

func writeCond(cond string, buf *strings.Builder) {
	cond = strings.TrimSpace(cond)
	if len(cond) < 1 {
		return
	}

	w := false
	ind := strings.IndexRune(cond, ' ')
	if ind > 0 {
		w = strings.ToLower(cond[:ind]) == "where"
	}

	if w {
		buf.WriteByte(' ')
		buf.WriteString(cond)
	} else {
		buf.WriteString(" WHERE ")
		buf.WriteString(cond)
	}
}

var quoteIdentifier = true

func DisableQuoteIdentifier() { quoteIdentifier = false }

func writeIdentifier(driver, v string, buf *strings.Builder) {
	if !quoteIdentifier {
		buf.WriteString(v)
		return
	}

	mysql := func(a string) {
		if strings.ContainsRune(a, '`') {
			a = strings.ReplaceAll(a, "`", "\\`")
		}
		buf.WriteRune('`')
		buf.WriteString(a)
		buf.WriteRune('`')
	}

	postgres := func(a string) {
		if strings.ContainsRune(a, '"') {
			a = strings.ReplaceAll(a, "\"", "\\\"")
		}
		buf.WriteRune('"')
		buf.WriteString(a)
		buf.WriteRune('"')
	}

	vl := utils.SplitAndTrim(v, ".")
	switch driver {
	case "mysql":
		for i, b := range vl {
			mysql(b)
			if i < len(vl)-1 {
				buf.WriteRune('.')
			}
		}
	case "postgres":
		for i, b := range vl {
			postgres(b)
			if i < len(vl)-1 {
				buf.WriteRune('.')
			}
		}
	default:
		buf.WriteString(v)
	}
}

func (op *Operator) simpleUpdate(driver, cond string, m Data) (string, Data) {
	if len(cond) < 1 || len(m) < 1 {
		panic(ErrEmptyConditionOrEmptyData)
	}

	buf := strings.Builder{}
	buf.WriteString("UPDATE ")
	writeIdentifier(driver, op.TableName(), &buf)
	buf.WriteString(" SET ")

	var retM Data
	i := 0
	lastInd := len(m) - 1
	for k, v := range m {
		if op.IsImmutableField(k) {
			panic(ErrImmutableField)
		}

		writeIdentifier(driver, k, &buf)
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
	writeCond(cond, &buf)
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
	exe := Exe(ctx)
	return exe.Get(ctx, op.simpleSelect(exe.Executor.DriverName(), keysGroup, condition), namedargs, dist)
}

func joinColumns(d string, v []string, buf *strings.Builder) {
	for i, k := range v {
		writeIdentifier(d, k, buf)
		if i < len(v)-1 {
			buf.WriteRune(',')
		}
	}
}

func (op *Operator) RowColumns(ctx context.Context, columns []string, cond string, namedargs interface{}, dist ...interface{}) error {
	exe := Exe(ctx)

	buf := strings.Builder{}
	buf.WriteString("SELECT ")
	joinColumns(exe.Executor.DriverName(), columns, &buf)
	buf.WriteString(" FROM ")
	buf.WriteString(op.TableName())
	writeIdentifier(exe.Executor.DriverName(), op.TableName(), &buf)
	buf.WriteRune(' ')
	writeCond(cond, &buf)
	return exe.ScanRow(ctx, buf.String(), namedargs, dist...)
}

func (op *Operator) RowsColumn(ctx context.Context, column string, cond string, namedargs interface{}, dist interface{}) error {
	exe := Exe(ctx)

	buf := strings.Builder{}
	buf.WriteString("SELECT ")
	writeIdentifier(exe.Executor.DriverName(), column, &buf)
	buf.WriteString(" FROM ")
	writeIdentifier(op.TableName(), column, &buf)
	buf.WriteRune(' ')
	writeCond(cond, &buf)
	return exe.Select(ctx, buf.String(), namedargs, dist)
}

func (op *Operator) FetchMany(ctx context.Context, keysGroup string, condition string, arg interface{}, dist interface{}) error {
	exe := Exe(ctx)
	return exe.Select(ctx, op.simpleSelect(exe.Executor.DriverName(), keysGroup, condition), arg, dist)
}

type Raw string
type Data map[string]interface{}

func (op *Operator) Update(ctx context.Context, data Data, condition string, namedargs interface{}) (sql.Result, error) {
	exe := Exe(ctx)
	s, m := op.simpleUpdate(exe.Executor.DriverName(), condition, data)
	return exe.Exec(ctx, s, mergeMap(ctx, m, namedargs))
}

func (op *Operator) simpleInsert(driver string, data Data, returning string) (string, Data) {
	if len(data) < 1 {
		panic(ErrEmptyConditionOrEmptyData)
	}

	buf := strings.Builder{}
	buf.WriteString("INSERT INTO ")
	writeIdentifier(driver, op.TableName(), &buf)
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
		writeIdentifier(driver, a, &buf)
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

func (op *Operator) Insert(ctx context.Context, data Data) (int64, error) {
	exe := Exe(ctx)
	s, m := op.simpleInsert(exe.Executor.DriverName(), data, "")
	r, e := exe.Exec(ctx, s, m)
	if e != nil {
		return 0, e
	}
	return r.LastInsertId()
}

func (op *Operator) InsertWithReturning(ctx context.Context, data Data, returning string, dist ...interface{}) {
	exe := Exe(ctx)
	s, m := op.simpleInsert(exe.Executor.DriverName(), data, returning)
	if err := exe.ScanRow(ctx, s, m, dist); err != nil {
		panic(err)
	}
}

var ErrDeleteWithoutCondition = errors.New("sha.sqlx: delete without condition")

func (op *Operator) Delete(ctx context.Context, cond string, namedargs interface{}) (sql.Result, error) {
	if len(cond) < 1 {
		panic(ErrDeleteWithoutCondition)
	}

	exe := Exe(ctx)

	var buf strings.Builder
	buf.WriteString("DELETE FROM ")
	writeIdentifier(exe.Executor.DriverName(), op.TableName(), &buf)
	writeCond(cond, &buf)
	return exe.Exec(ctx, buf.String(), namedargs)
}
