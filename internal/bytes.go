// source https://github.com/savsgio/gotils/blob/master/conv.go

package internal

import (
	"reflect"
	"unsafe"
)

// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
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

var spaceMap []byte

func init() {
	spaceMap = make([]byte, 128)
	for i := 0; i < 128; i++ {
		if i <= 32 || i == 127 {
			spaceMap[i] = 1
		} else {
			spaceMap[i] = 0
		}
	}
}

func InplaceTrimAsciiSpace(v []byte) []byte {
	var left = 0
	var right = len(v) - 1
	for ; left <= right; left++ {
		b := v[left]
		if b > 127 {
			break
		}
		if spaceMap[b] != 1 {
			break
		}
	}
	for ; right >= left; right-- {
		b := v[left]
		if b > 127 {
			break
		}
		if spaceMap[b] != 1 {
			break
		}
	}
	return v[left : right+1]
}
