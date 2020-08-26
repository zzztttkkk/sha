package grafql

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/zzztttkkk/suna/reflectx"
	"reflect"
	"strings"
	"sync"
)

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
	return v.getTag(field) != "-"
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
	var f *graphql.Field
	ele := reflect.New(field.Type)

	switch ele.Interface().(type) {
	case string, *string:
		f = &graphql.Field{Type: graphql.String}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64,
		*int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
		f = &graphql.Field{Type: graphql.Int}
	case float32, float64, *float32, *float64:
		f = &graphql.Field{Type: graphql.Float}
	case bool, *bool:
		f = &graphql.Field{Type: graphql.Boolean}
	case []byte, *[]byte:
		f = &graphql.Field{Type: graphql.String}
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
	fields, ok := _ReflectOutCache.Load(t)
	if ok {
		return graphql.NewObject(
			graphql.ObjectConfig{
				Name:   name,
				Fields: fields.(map[string]*graphql.Field),
			},
		)
	}

	visitor := _OutVisitor{fields: map[string]*graphql.Field{}, names: map[string]string{}}
	reflectx.Map(t, &visitor)
	_ReflectOutCache.Store(t, visitor.fields)

	for k, field := range visitor.fields {
		fn := fmt.Sprintf("Resolve%s", visitor.names[k])
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

	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   name,
			Fields: visitor.fields,
		},
	)
}
