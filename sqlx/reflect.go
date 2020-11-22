package sqlx

import (
	"database/sql"
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/internal/typereflect"
	"reflect"
	"time"
)

type Field struct {
	Name  string
	Index []int
	IsPtr bool
}

type Fields []*Field

type _TagParser struct {
	current *Field
	field   *reflect.StructField
	fields  Fields
	name    string
}

var timeType = reflect.TypeOf(time.Time{})

func (p *_TagParser) OnNestedStruct(f *reflect.StructField, index []int) typereflect.OnNestStructRet {
	if !f.Anonymous {
		if f.Type == timeType {
			return typereflect.None
		}

		ele := reflect.New(f.Type).Interface()
		if _, ok := ele.(sql.Scanner); ok {
			return typereflect.None
		}
		return typereflect.Skip
	}
	return typereflect.GoDown
}

func (p *_TagParser) OnBegin(f *reflect.StructField, index []int) bool {
	p.field = f
	if f.Type.Kind() == reflect.Ptr {
		internal.L.Printf("suna.sqlx: skip pointer field\n")
		return false
	}

	p.current = &Field{}
	rule := p.current
	rule.Index = append(rule.Index, index...)

	if f.Type.Kind() == reflect.Struct {
		if f.Type == timeType {
			return true
		}

		ele := reflect.New(f.Type).Interface()
		if _, ok := ele.(sql.Scanner); ok {
			return true
		}

		p.fields = append(p.fields, GetFields(f.Type)...)
		return false
	}

	return true
}

func (p *_TagParser) OnName(name string) {
	p.current.Name = name
}

func (p *_TagParser) OnAttr(key, val string) {
	switch key {

	}
}

func (p *_TagParser) OnDone() {
	field := p.current
	p.fields = append(p.fields, field)

	p.current = nil
	p.field = nil
}

var m = map[reflect.Type]Fields{}

func GetFields(t reflect.Type) Fields {
	fs := m[t]
	if fs != nil {
		return fs
	}

	p := _TagParser{}
	p.name = fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
	typereflect.Tags(t, "sqlx", &p)
	m[t] = p.fields
	return p.fields
}
