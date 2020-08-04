package reflectx

import (
	"reflect"
	"strings"
)

type _VisitorT struct {
	onNS func(field *reflect.StructField) bool
	onF  func(field *reflect.StructField)
}

func (v *_VisitorT) OnNestStructField(field *reflect.StructField) bool {
	return v.onNS(field)
}

func (v *_VisitorT) OnField(field *reflect.StructField) {
	v.onF(field)
}

func Keys(rt reflect.Type, tag string) (lst []string) {
	Map(
		rt,
		&_VisitorT{
			onNS: func(field *reflect.StructField) bool {
				return field.Tag.Get(tag) != "-"
			},
			onF: func(field *reflect.StructField) {
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
	)
	return
}

func ExportedKeys(rt reflect.Type, tag string) (lst []string) {
	Map(
		rt,
		&_VisitorT{
			onNS: func(field *reflect.StructField) bool {
				return field.Tag.Get(tag) != "-"
			},
			onF: func(field *reflect.StructField) {
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
	)
	return
}
