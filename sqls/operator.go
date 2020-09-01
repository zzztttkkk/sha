package sqls

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"reflect"
	"strings"
)

type Operator struct {
	tablename   string
	idField     string
	ele         reflect.Value
	dbGroupName string
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

func (op *Operator) _GetDbLeader() *sqlx.DB {
	var db = cfg.GetSqlLeader()
	if len(op.dbGroupName) > 0 {
		dbs, ok := _DbGroups[op.dbGroupName]
		if !ok {
			panic(fmt.Errorf("suna.sqls: database group `%s` is not exists", op.dbGroupName))
		}
		db = dbs.Leader
	}
	return db
}

func (op *Operator) Init(v interface{}) {
	op.ele = reflect.ValueOf(v)

	fnV := op.ele.MethodByName("DatabaseGroup")
	if fnV.IsValid() {
		op.dbGroupName = fnV.Call(nil)[0].Interface().(string)
	}

	CreateTable(op._GetDbLeader(), op.ele, op.TableName())
}

func getTableName(ele reflect.Value) string {
	tablenameFn := ele.MethodByName("TableName")
	if tablenameFn.IsValid() {
		return (tablenameFn.Call(nil)[0]).Interface().(string)
	}
	return strings.ToLower(ele.Type().Name())
}

func CreateTable(db *sqlx.DB, ele reflect.Value, name string) {
	fnV := ele.MethodByName("TableDefinition")
	fnT := fnV.Type()
	msg := fmt.Sprintf(
		"suna.sqls: `%s.%s` has no method `TableDefinition`\n",
		ele.Type().PkgPath(), ele.Type().Name(),
	)
	if !fnV.IsValid() || fnT.NumOut() != 1 {
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
	case 1: // TableDefinition(*sqlx.DB)
		switch reflect.New(fnT.In(0)).Elem().Interface().(type) {
		case *sqlx.DB:
		default:
			log.Print(msg)
			return
		}
		fields = (fnV.Call([]reflect.Value{reflect.ValueOf(db)})[0]).Interface().([]string)
	case 0: // TableDefinition(*sqlx.DB)
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
	db.MustExec(q)
}

func (op *Operator) BindNamed(q string, m map[string]interface{}) (string, []interface{}) {
	s, vl, err := op._GetDbLeader().BindNamed(q, m)
	if err != nil {
		panic(err)
	}
	return s, vl
}
