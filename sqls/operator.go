package sqls

import (
	"fmt"
	"reflect"
	"strings"
)

type Operator struct {
	tablename string
	idField   string
}

func (op *Operator) TableName() string {
	return op.tablename
}

func (op *Operator) SetIdField(f string) {
	op.idField = strings.TrimSpace(strings.Split(strings.TrimSpace(f), ",")[0])
}

func (op *Operator) Init(ele reflect.Value) {
	op.tablename = getTableName(ele)
	CreateTable(ele)
}

func getTableName(ele reflect.Value) string {
	tablenameFn := ele.MethodByName("TableName")
	if tablenameFn.IsValid() {
		return (tablenameFn.Call(nil)[0]).Interface().(string)
	}
	return strings.ToLower(ele.Type().Name())
}

func CreateTable(ele reflect.Value) {
	tablecreationFn := ele.MethodByName("TableDefinition")
	if !tablecreationFn.IsValid() {
		return
	}

	lines := (tablecreationFn.Call(nil)[0]).Interface().([]string)
	q := fmt.Sprintf(
		"create table if not exists %s (%s)",
		getTableName(ele),
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
