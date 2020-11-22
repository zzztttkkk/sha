package typereflect

import (
	"reflect"
)

type OnNestStructRet int

const (
	None = OnNestStructRet(iota)
	Skip
	GoDown
)

type Visitor interface {
	// return true, if need go down
	OnNestStruct(field *reflect.StructField, index []int) OnNestStructRet
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

			switch visitor.OnNestStruct(&field, nIndex) {
			case None:
				visitor.OnField(&field, copyAppend(index, i))
			case GoDown:
				Map(field.Type, visitor, nIndex)
			}
			continue
		}

		visitor.OnField(&field, copyAppend(index, i))
	}
}
