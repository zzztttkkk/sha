package typereflect

import (
	"reflect"
)

type Visitor interface {
	// return true, if need go down
	OnNestStruct(field *reflect.StructField, index []int) bool
	OnField(field *reflect.StructField, index []int)
}

func copyAppend(v []int, i int) []int {
	var rv []int
	rv = append(rv, v...)
	rv = append(rv, i)
	return rv
}

func Map(t reflect.Type, visitor Visitor, index []int) {
	num := t.NumField()
	for i := 0; i < num; i++ {
		field := t.Field(i)
		if field.Type.Kind() == reflect.Struct {
			nIndex := copyAppend(index, i)
			if visitor.OnNestStruct(&field, nIndex) {
				Map(field.Type, visitor, nIndex)
			}
			continue
		}

		visitor.OnField(&field, copyAppend(index, i))
	}
}
