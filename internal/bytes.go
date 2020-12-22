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

var spaceMap [128]bool

func init() {
	for i := 0; i < 128; i++ {
		switch i {
		case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0:
			spaceMap[i] = true
		default:
			spaceMap[i] = false
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
		if !spaceMap[b] {
			break
		}
	}
	for ; right >= left; right-- {
		b := v[right]
		if b > 127 {
			break
		}
		if !spaceMap[b] {
			break
		}
	}
	return v[left : right+1]
}
