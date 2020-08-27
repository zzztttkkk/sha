package grafql

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/zzztttkkk/suna/reflectx"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Outer interface {
	GraphqlScalar() *graphql.Scalar
}

type _OutVisitor struct {
	fields map[string]*graphql.Field
	names  map[string]string
}

func (v *_OutVisitor) getTag(f *reflect.StructField) string {
	jt := f.Tag.Get("json")
	if len(jt) > 0 {
		return jt
	}
	gt := f.Tag.Get("graphql")
	if len(gt) > 0 {
		return gt
	}
	return ""
}

func (v *_OutVisitor) OnNestStructField(field *reflect.StructField) bool {
	t := v.getTag(field)
	if t == "-" {
		return false
	}

	ele := reflect.New(field.Type)
	switch rv := ele.Interface().(type) {
	case Outer:
		var name string
		if len(t) < 1 {
			name = strings.ToLower(field.Name)
		} else {
			name = strings.TrimSpace(strings.Split(t, ",")[0])
		}
		name = strings.TrimLeft(name, "_")

		f := &graphql.Field{Type: rv.GraphqlScalar()}
		f.Name = strings.TrimLeft(field.Name, "_")
		v.fields[name] = f
		return false
	}
	return true
}

func (v *_OutVisitor) OnField(field *reflect.StructField) {
	t := v.getTag(field)
	if t == "-" {
		return
	}

	var name string
	if len(t) < 1 {
		name = strings.ToLower(field.Name)
	} else {
		name = strings.TrimSpace(strings.Split(t, ",")[0])
	}
	name = strings.TrimLeft(name, "_")

	var f *graphql.Field
	ele := reflect.New(field.Type)

	switch rv := ele.Interface().(type) {
	case string, *string, []byte, *[]byte:
		f = &graphql.Field{Type: graphql.String}
	case []string, [][]byte:
		f = &graphql.Field{Type: graphql.NewList(graphql.String)}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64,
		*int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
		f = &graphql.Field{Type: graphql.Int}
	case []int, []int8, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64:
		f = &graphql.Field{Type: graphql.NewList(graphql.Int)}
	case float32, float64, *float32, *float64:
		f = &graphql.Field{Type: graphql.Float}
	case []float32, []float64:
		f = &graphql.Field{Type: graphql.NewList(graphql.Float)}
	case bool, *bool:
		f = &graphql.Field{Type: graphql.Boolean}
	case []bool:
		f = &graphql.Field{Type: graphql.NewList(graphql.Boolean)}
	case time.Time, *time.Time:
		f = &graphql.Field{Type: graphql.DateTime}
	case []time.Time, []*time.Time:
		f = &graphql.Field{Type: graphql.NewList(graphql.DateTime)}
	case Outer:
		f = &graphql.Field{Type: rv.GraphqlScalar()}
	default:
		return
	}

	f.Name = name
	v.fields[name] = f
	v.names[name] = field.Name
}

var _ReflectOutCache sync.Map

func NewOutObject(v reflect.Value) *graphql.Object {
	t := v.Type()

	name := strings.TrimLeft(t.Name(), "_")
	obj, ok := _ReflectOutCache.Load(t)
	if ok {
		return obj.(*graphql.Object)
	}

	visitor := _OutVisitor{fields: map[string]*graphql.Field{}, names: map[string]string{}}
	reflectx.Map(t, &visitor)

	for k, field := range visitor.fields {
		_RawFieldName := visitor.names[k]
		if len(_RawFieldName) < 1 {
			// custom scalar type
			continue
		}

		fn := fmt.Sprintf("ResolveOut%s", _RawFieldName)
		fv := v.MethodByName(fn)
		if !fv.IsValid() {
			continue
		}

		rf, ok := fv.Interface().(func(params graphql.ResolveParams) (interface{}, error))
		if ok {
			field.Resolve = rf
			continue
		}
	}

	rv := graphql.NewObject(
		graphql.ObjectConfig{
			Name:   name,
			Fields: visitor.fields,
		},
	)
	_ReflectOutCache.Store(t, rv)
	return rv
}
