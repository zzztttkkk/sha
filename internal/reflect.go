package internal

import (
	"reflect"
	"unsafe"
)

func ChangeField(dist interface{}, fn string, fv interface{}) bool {
	v := reflect.ValueOf(dist)
	if v.Kind() != reflect.Ptr {
		return false
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return false
	}

	f := v.FieldByName(fn)
	if !f.IsValid() {
		return false
	}
	if fn[0] >= 'A' && fn[0] <= 'Z' && f.CanSet() {
		f.Set(reflect.ValueOf(fv))
		return true
	}
	reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Set(reflect.ValueOf(fv))
	return true
}
