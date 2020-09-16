package sqls

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Operator struct {
	tablename string
	ele       reflect.Value
}

func (op *Operator) TableName() string {
	if len(op.tablename) > 0 {
		return op.tablename
	}
	op.tablename = getTableName(op.ele)
	return op.tablename
}

func (op *Operator) Init(v interface{}) {
	op.ele = reflect.ValueOf(v)
	CreateTable(cfg.GetSqlLeader(), op.ele, op.TableName())
}

func getTableName(ele reflect.Value) string {
	tablenameFn := ele.MethodByName("SqlsTableName")
	if tablenameFn.IsValid() {
		return (tablenameFn.Call(nil)[0]).Interface().(string)
	}
	return strings.ToLower(ele.Type().Name())
}

func CreateTable(db *sqlx.DB, ele reflect.Value, name string) {
	fnV := ele.MethodByName("SqlsTableColumns")
	msg := fmt.Sprintf(
		"suna.sqls: `%s.%s` has no method `SqlsTableColumns`, or type error\n",
		ele.Type().PkgPath(), ele.Type().Name(),
	)

	if !fnV.IsValid() {
		log.Print(msg)
		return
	}
	fnT := fnV.Type()
	if fnT.NumOut() != 1 {
		log.Print(msg)
		return
	}
	switch reflect.New(fnT.Out(0)).Elem().Interface().(type) {
	case []string:
	default:
		log.Print(msg)
		return
	}

	var fields []string
	switch fnV.Type().NumIn() {
	case 1: // SqlsTableColumns(*sqlx.DB)
		switch reflect.New(fnT.In(0)).Elem().Interface().(type) {
		case *sqlx.DB:
		default:
			log.Print(msg)
			return
		}
		fields = (fnV.Call([]reflect.Value{reflect.ValueOf(db)})[0]).Interface().([]string)
	case 0: // SqlsTableColumns()
		fields = (fnV.Call(nil)[0]).Interface().([]string)
	default:
		log.Print(msg)
		return
	}

	q := fmt.Sprintf(
		"create table if not exists %s (%s)",
		name,
		strings.Join(fields, ","),
	)
	_DoSqlLogging(q, nil)
	db.MustExec(q)
}
