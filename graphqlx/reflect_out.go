package graphqlx

import (
	"context"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/zzztttkkk/suna/reflectx"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"
)

type GraphqlScalarer interface {
	GraphqlScalar() *graphql.Scalar
}

type ResolveOutFunction func(ctx context.Context, info *graphql.ResolveInfo) (interface{}, error)

type _OutVisitor struct {
	fields  graphql.Fields
	name    string
	current reflect.Type
}

func (v *_OutVisitor) getTag(f *reflect.StructField) string {
	return f.Tag.Get("json")
}

func (v *_OutVisitor) OnNestStructField(field *reflect.StructField) bool {
	t := v.getTag(field)
	if t == "-" {
		return false
	}

	ele := reflect.New(field.Type).Elem()
	switch rv := ele.Interface().(type) {
	case GraphqlScalarer:
		if field.Anonymous {
			panic(
				fmt.Errorf(
					"suna.graphqlx: interface `GraphqlScalarer` can not be inherited; %s.%s",
					v.name, field.Name,
				),
			)
		}

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
	ele := reflect.New(field.Type).Elem()

	switch rv := ele.Interface().(type) {
	case string:
		f = &graphql.Field{Type: graphql.String}
	case []string:
		f = &graphql.Field{Type: graphql.NewList(graphql.String)}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		f = &graphql.Field{Type: graphql.Int}
	case []int, []int8, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64:
		f = &graphql.Field{Type: graphql.NewList(graphql.Int)}
	case float32, float64:
		f = &graphql.Field{Type: graphql.Float}
	case []float32, []float64:
		f = &graphql.Field{Type: graphql.NewList(graphql.Float)}
	case bool:
		f = &graphql.Field{Type: graphql.Boolean}
	case []bool:
		f = &graphql.Field{Type: graphql.NewList(graphql.Boolean)}
	case time.Time, *time.Time:
		f = &graphql.Field{Type: graphql.DateTime}
	case []time.Time, []*time.Time:
		f = &graphql.Field{Type: graphql.NewList(graphql.DateTime)}
	case GraphqlScalarer:
		f = &graphql.Field{Type: rv.GraphqlScalar()}
	default:
		// handle []CustomStruct/[]*CustomStruct
		if field.Type.Kind() == reflect.Slice {
			var ele reflect.Value
			eleT := field.Type.Elem()
			if eleT.Kind() == reflect.Struct {
				if eleT == v.current {
					panic(fmt.Errorf("suna.graphqlx: circular reference, %s.%s", v.name, field.Name))
				}
				ele = reflect.New(eleT).Elem() // []A
			} else if eleT.Kind() == reflect.Ptr {
				eleT = eleT.Elem()
				if eleT.Kind() == reflect.Struct {
					if eleT == v.current {
						panic(fmt.Errorf("suna.graphqlx: circular reference, %s.%s", v.name, field.Name))
					}
					ele = reflect.New(eleT).Elem() // []*A
				}
			}
			if ele.IsValid() {
				switch rv := ele.Interface().(type) {
				case GraphqlScalarer:
					var name string
					if len(t) < 1 {
						name = strings.ToLower(field.Name)
					} else {
						name = strings.TrimSpace(strings.Split(t, ",")[0])
					}
					name = strings.TrimLeft(name, "_")
					f = &graphql.Field{Type: graphql.NewList(rv.GraphqlScalar())}
				default:
					f = &graphql.Field{Type: graphql.NewList(NewOutObjectType(ele))}
				}
			}
		}
	}

	if f == nil {
		log.Printf("suna.graphqlx: unqupported type, %s.%s", v.name, field.Name)
		return
	}

	f.Name = name
	v.fields[name] = f
}

var _ReflectOutCache sync.Map

func getOutputFields(t reflect.Type) *graphql.ObjectConfig {
	name := strings.TrimLeft(t.Name(), "_")
	visitor := _OutVisitor{
		fields:  map[string]*graphql.Field{},
		name:    fmt.Sprintf("%s.%s", t.PkgPath(), t.Name()),
		current: t,
	}
	reflectx.Map(t, &visitor)
	rv := &graphql.ObjectConfig{
		Name:   name,
		Fields: visitor.fields,
	}
	return rv
}

func NewOutObjectType(v reflect.Value) graphql.Output {
	t := v.Type()
	o, ok := _ReflectOutCache.Load(t)
	if ok {
		return o.(graphql.Output)
	}

	if t.Kind() == reflect.Slice {
		et := t.Elem()
		if et.Kind() != reflect.Struct {
			panic("suna.grfqlx: not a struct slice")
		}
		rv := graphql.NewList(graphql.NewObject(*getOutputFields(et)))
		_ReflectOutCache.Store(t, rv)
		return rv
	}
	rv := graphql.NewObject(*getOutputFields(t))
	_ReflectOutCache.Store(t, rv)
	return rv
}
