package sqls

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/zzztttkkk/snow/reflectx"

	"github.com/zzztttkkk/snow/utils"
)

type _SqlFieldT struct {
	name         string
	isPrimary    bool
	notNull      bool
	autoIncr     bool
	unique       bool
	length       int
	defaultValue string
	storageType  string
	isUnsigned   bool
}

type _SqlFieldSlice []*_SqlFieldT

type _DdlParser struct {
	current   *_SqlFieldT
	fields    _SqlFieldSlice
	primaries []string
	tableName string
}

var sqlNullStringType = reflect.TypeOf(sql.NullString{})

func (p *_DdlParser) OnNestStruct(field *reflect.StructField) bool {
	if field.Tag.Get("ddl") == "-" {
		return false
	}

	if field.Type == sqlNullStringType {
		return false
	}

	return true
}

func (p *_DdlParser) OnBegin(field *reflect.StructField) bool {
	mf := &_SqlFieldT{}

	switch field.Type.Kind() {
	case reflect.Bool:
		mf.storageType = "bool"
	case reflect.Int8:
		mf.storageType = "tinyint"
	case reflect.Int16:
		mf.storageType = "smallint"
	case reflect.Int, reflect.Int32:
		mf.storageType = "int"
	case reflect.Int64:
		mf.storageType = "bigint"
	case reflect.Uint8:
		mf.storageType = "tinyint"
		mf.isUnsigned = true
	case reflect.Uint16:
		mf.storageType = "smallint"
		mf.isUnsigned = true
	case reflect.Uint, reflect.Uint32:
		mf.isUnsigned = true
		mf.storageType = "int"
	case reflect.Uint64:
		mf.isUnsigned = true
		mf.storageType = "bigint"
	case reflect.Slice:
		switch field.Type.Elem().Kind() {
		case reflect.Interface: // []interface{}, utils.SqlArray
			mf.storageType = "json"
		case reflect.Uint8: // []byte
			mf.storageType = "string"
		default:
			return false
		}
	case reflect.Map:
		if field.Type.Key().Kind() == reflect.String && field.Type.Elem().Kind() == reflect.Interface {
			mf.storageType = "json"
		} else {
			return false
		}
	case reflect.String:
		mf.storageType = "string"
	default:
		switch field.Type {
		case sqlNullStringType:
			mf.storageType = "string"
		default:
			return false
		}
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
		mf.length = int(utils.S2I32(val))
	case "D", "default":
		if len(val) > 0 {
			mf.defaultValue = val
		}
	case "T", "type":
		mf.storageType = val
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
	MaxCharLength = 128
)

var ddlCache = map[reflect.Type]*_DdlParser{}

func newDdlParser(p reflect.Type) *_DdlParser {
	mfs, ok := ddlCache[p]
	if ok {
		return mfs
	}
	parser := &_DdlParser{}
	reflectx.Tags(p, "ddl", parser)
	ddlCache[p] = parser

	tableName := strings.ToLower(p.Name())
	ele := reflect.New(p)
	tableNameFn := ele.MethodByName("TableName")
	if tableNameFn.IsValid() {
		tableName = (tableNameFn.Call(nil)[0]).Interface().(string)
	}
	parser.tableName = tableName
	return parser
}

// support mysql types: bool, tinyint, smallint, int, bigint, char, varchar, text, blob, json
//
// constraints: primary, notnull, unique, incr, const
//
// string option: L<length>:
// 				-1: blob,
// 				0: text,
// 				1-MaxCharLength : char,
// 				MaxCharLength~: varchar,
// default options: D<default value>
//
// example:
// type User {
//     Identify 		uint32  `ddl:"primary;notnull;incr;"`
// 	   GetName 	string  `ddl:"unique;notnull;unique;L<50>"`
// 	   Address 	string  `ddl:"addr:notnull;L<120>"`
// 	   Age		int		`ddl:"notnull"`
// }
func TableDefinition(modelType reflect.Type) string {
	dn := string(config.GetMust("sql.driver"))
	if dn != "mysql" {
		panic(fmt.Errorf("snow.sqls: unsuport driver `%s`", dn))
	}

	ele := reflect.New(modelType)
	cFn := ele.MethodByName("TableCreation")
	if cFn.IsValid() {
		return (cFn.Call(nil)[0]).Interface().(string)
	}

	parser := newDdlParser(modelType)
	buf := strings.Builder{}
	buf.WriteString("create table if not exists " + parser.tableName + " (\n")

	for _, field := range parser.fields {
		buf.WriteString("\t")
		buf.WriteString(field.name)
		buf.WriteString(" ")
		if field.storageType == "string" {
			if field.length < 0 {
				buf.WriteString("blob")
			} else if field.length == 0 {
				buf.WriteString("text")
			} else if field.length <= MaxCharLength {
				buf.WriteString(fmt.Sprintf("char(%d)", field.length))
			} else {
				buf.WriteString(fmt.Sprintf("varchar(%d)", field.length))
			}
		} else {
			buf.WriteString(field.storageType)
		}

		if field.isUnsigned {
			buf.WriteString(" unsigned")
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
