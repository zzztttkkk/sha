package sqls

import (
	"fmt"
	"github.com/zzztttkkk/snow/internal"
	"reflect"
	"strings"
)

type _MysqlFieldT struct {
	name         string
	t            string
	isPrimary    bool
	notNull      bool
	autoIncr     bool
	unique       bool
	length       int
	defaultValue string
	isConst      bool
}

type _DdlParser struct {
	current   *_MysqlFieldT
	fields    []*_MysqlFieldT
	primaries []string
}

func (p *_DdlParser) Tag() string {
	return "ddl"
}

func (p *_DdlParser) OnField(field *reflect.StructField) bool {
	mf := &_MysqlFieldT{}

	switch field.Type.Kind() {
	case reflect.Bool:
		mf.t = "bool"
	case reflect.Int8:
		mf.t = "tinyint"
	case reflect.Int16:
		mf.t = "smallint"
	case reflect.Int, reflect.Int32:
		mf.t = "int"
	case reflect.Int64:
		mf.t = "bigint"
	case reflect.Uint8:
		mf.t = "tinyint unsigned"
	case reflect.Uint16:
		mf.t = "smallint unsigned"
	case reflect.Uint, reflect.Uint32:
		mf.t = "int unsigned"
	case reflect.Uint64:
		mf.t = "bigint unsigned"
	case reflect.Slice:
		switch field.Type.Elem().Kind() {
		case reflect.Interface: // []interface{}, utils.SqlArray
			mf.t = "json"
		case reflect.Uint8: // []byte
			mf.t = "string"
		default:
			return false
		}
	case reflect.Map:
		if field.Type.Key().Kind() == reflect.String && field.Type.Elem().Kind() == reflect.Interface {
			mf.t = "json"
		} else {
			return false
		}
	case reflect.String:
		mf.t = "string"
	default:
		return false
	}
	p.current = mf
	return true
}

func (p *_DdlParser) OnName(name string) {
	p.current.name = name
}

func (p *_DdlParser) OnAttr(key, val string) {
	mf := p.current
	switch key {
	case "primary":
		mf.isPrimary = true
	case "notnull":
		mf.notNull = true
	case "incr":
		mf.autoIncr = true
	case "unique":
		mf.unique = true
	case "L", "length":
		mf.length = int(internal.S2I32(val))
	case "D", "default":
		if len(val) > 0 {
			mf.defaultValue = val
		}
	}
}

func (p *_DdlParser) OnDone() {
	p.fields = append(p.fields, p.current)
	if p.current.isPrimary {
		p.primaries = append(p.primaries, p.current.name)
	}
	p.current = nil
}

const (
	MysqlMaxCharLength = 100
)

// support mysql types: bool, tinyint, smallint, int, bigint, char, varchar, text, blob, json
//
// constraints: primary, notnull, unique, incr, const
//
// string option: L<length>:
//				-1: blob,
//				0: text,
//				1-MysqlMaxCharLength : char,
//				MysqlMaxCharLength~: varchar,
// default options: D<default value>
//
// example:
// type User {
//     Identify 		uint32  `ddl:"primary;notnull;incr;"`
// 	   Name 	string  `ddl:"unique;notnull;unique;L<50>"`
//	   Address 	string  `ddl:"addr:notnull;L<120>"`
//	   Age		int		`ddl:"notnull"`
// }
func TableDefinition(modelType reflect.Type) string {
	ele := reflect.New(modelType)
	cFn := ele.MethodByName("TableCreation")
	if cFn.IsValid() {
		return (cFn.Call(nil)[0]).Interface().(string)
	}

	parser := _DdlParser{}
	internal.ReflectTags(modelType, &parser)

	buf := strings.Builder{}
	tableName := strings.ToLower(modelType.Name())
	tableNameFn := ele.MethodByName("TableName")
	if tableNameFn.IsValid() {
		tableName = (tableNameFn.Call(nil)[0]).Interface().(string)
	}

	buf.WriteString("create table if not exists " + tableName + " (\n")

	for _, field := range parser.fields {
		buf.WriteString("\t")
		buf.WriteString(field.name)
		buf.WriteString(" ")
		if field.t == "string" {
			if field.length < 0 {
				buf.WriteString("blob")
			} else if field.length == 0 {
				buf.WriteString("text")
			} else if field.length <= MysqlMaxCharLength {
				buf.WriteString(fmt.Sprintf("char(%d)", field.length))
			} else {
				buf.WriteString(fmt.Sprintf("varchar(%d)", field.length))
			}
		} else {
			buf.WriteString(field.t)
		}

		if field.autoIncr {
			buf.WriteString(" auto_increment")
		}

		if field.unique {
			buf.WriteString(" unique")
		}

		if len(field.defaultValue) > 0 {
			buf.WriteString(" default " + field.defaultValue)
		}

		buf.WriteString(",")
		buf.WriteString("\n")
	}

	buf.WriteString(fmt.Sprintf("\tprimary key(%s)", strings.Join(parser.primaries, ",")))
	buf.WriteString("\n);\n")
	return buf.String()
}
