package typereflect

import (
	"reflect"
)

type _Visitor struct {
	ns func(field *reflect.StructField, index []int) OnNestStructRet
	f  func(field *reflect.StructField, index []int)
}

func (v *_Visitor) OnNestStruct(field *reflect.StructField, index []int) OnNestStructRet {
	return v.ns(field, index)
}

func (v *_Visitor) OnField(field *reflect.StructField, index []int) {
	v.f(field, index)
}
