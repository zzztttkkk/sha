// source https://github.com/savsgio/gotils/blob/master/conv.go

package utils

import (
	"reflect"
	"unsafe"
)

// S See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
func S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func B(s string) (b []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return
}
