package sqls

import (
	"fmt"
	"reflect"
	"strings"
)

type Operator struct {
	tablename string
	idField   string
	ele       reflect.Value
}

func (op *Operator) TableName() string {
	if len(op.tablename) > 0 {
		return op.tablename
	}
	op.tablename = getTableName(op.ele)
	return op.tablename
}

func (op *Operator) SetIdField(f string) {
	op.idField = strings.TrimSpace(strings.Split(strings.TrimSpace(f), ",")[0])
}

func (op *Operator) Init(ele reflect.Value) {
	op.ele = ele
	CreateTable(ele, op.TableName())
}

func getTableName(ele reflect.Value) string {
	tablenameFn := ele.MethodByName("TableName")
	if tablenameFn.IsValid() {
		return (tablenameFn.Call(nil)[0]).Interface().(string)
	}
	return strings.ToLower(ele.Type().Name())
}

func CreateTable(ele reflect.Value, name string) {
	tablecreationFn := ele.MethodByName("TableDefinition")
	if !tablecreationFn.IsValid() {
		return
	}

	lines := (tablecreationFn.Call(nil)[0]).Interface().([]string)
	q := fmt.Sprintf(
		"create table if not exists %s (%s)",
		name,
		strings.Join(lines, ","),
	)
	leader.MustExec(q)
}

func (op *Operator) BindNamed(q string, m map[string]interface{}) (string, []interface{}) {
	s, vl, err := leader.BindNamed(q, m)
	if err != nil {
		panic(err)
	}
	return s, vl
}
