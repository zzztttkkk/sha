package validator

import (
	"reflect"
)

type Field interface {
	FormValue(data []byte) bool
}

func (rule *_Rule) toCustomField(f *reflect.Value, data []byte) bool {
	t := rule.fieldType
	if rule.isPtr {
		t = t.Elem()
	}

	ptr := reflect.New(t)
	if ptr.Interface().(Field).FormValue(data) {
		if rule.isPtr {
			f.Set(ptr)
		} else {
			f.Set(ptr.Elem())
		}
		return true
	}
	return false
}
