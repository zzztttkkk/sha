package reflectx

import "reflect"

type Visitor interface {
	OnNestStruct(field *reflect.StructField) bool
	OnField(field *reflect.StructField)
}

func Map(t reflect.Type, visitor Visitor) {
	num := t.NumField()
	for i := 0; i < num; i++ {
		field := t.Field(i)

		if field.Type.Kind() == reflect.Struct {
			if visitor.OnNestStruct(&field) {
				Map(field.Type, visitor)
			}
			continue
		}

		visitor.OnField(&field)
	}
}
