package typereflect

import (
	"reflect"
	"strings"
)

type _Visitor struct {
	ns func(field *reflect.StructField, index []int) bool
	f  func(field *reflect.StructField, index []int)
}

func (v *_Visitor) OnNestStruct(field *reflect.StructField, index []int) bool {
	return v.ns(field, index)
}

func (v *_Visitor) OnField(field *reflect.StructField, index []int) {
	v.f(field, index)
}

func Keys(rt reflect.Type, tag string) (lst []string) {
	Map(
		rt,
		&_Visitor{
			ns: func(field *reflect.StructField, index []int) bool {
				return field.Tag.Get(tag) != "-"
			},
			f: func(field *reflect.StructField, index []int) {
				value := field.Tag.Get(tag)
				if value == "-" {
					return
				}
				ind := strings.IndexByte(value, ':')
				if ind < 1 {
					lst = append(lst, strings.ToLower(field.Name))
					return
				}
				lst = append(lst, value[:ind])
			},
		},
		nil,
	)
	return
}

func ExportedKeys(rt reflect.Type, tag string) (lst []string) {
	Map(
		rt,
		&_Visitor{
			ns: func(field *reflect.StructField, index []int) bool {
				return field.Tag.Get(tag) != "-"
			},
			f: func(field *reflect.StructField, index []int) {
				value := field.Tag.Get(tag)
				if value == "-" {
					return
				}
				if field.Name[0] >= 'A' && field.Name[0] <= 'Z' {
					ind := strings.IndexByte(value, ':')
					if ind < 1 {
						lst = append(lst, strings.ToLower(field.Name))
						return
					}
					lst = append(lst, value[:ind])
				}
			},
		},
		nil,
	)
	return
}
