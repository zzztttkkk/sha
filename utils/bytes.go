package utils

import (
	"reflect"
	"strconv"
	"sync"
	"unicode/utf8"
	"unsafe"
)

// https://github.com/savsgio/gotils/blob/master/conv.go#L13
func B2s(b []byte) string {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&b))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}

	return *(*string)(unsafe.Pointer(&bh))
}

// https://github.com/savsgio/gotils/blob/master/conv.go#28
func S2b(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}

	return *(*[]byte)(unsafe.Pointer(&bh))
}

func Runes(s []byte, count int) []rune {
	t := make([]rune, 0)
	i := 0
	for len(s) > 0 {
		r, l := utf8.DecodeRune(s)
		t = append(t, r)
		i++
		if i == count {
			break
		}
		s = s[l:]
	}
	return t
}

func S2U32(s string) uint32 {
	v, e := strconv.ParseUint(s, 10, 32)
	if e != nil {
		panic(e)
	}
	return uint32(v)
}

func B2U32(b []byte) uint32 {
	return S2U32(B2s(b))
}

func S2I32(s string) int32 {
	v, e := strconv.ParseInt(s, 10, 32)
	if e != nil {
		panic(e)
	}
	return int32(v)
}

func B2I32(b []byte) int32 {
	return S2I32(B2s(b))
}

type BytesPool struct {
	defaultSize int
	poll        sync.Pool
}

func NewBytesPool(cap, size int) *BytesPool {
	return &BytesPool{
		defaultSize: size,
		poll: sync.Pool{
			New: func() interface{} {
				v := make([]byte, size, cap)
				return &v
			},
		},
	}
}

func (bp *BytesPool) Get() *[]byte {
	return bp.poll.Get().(*[]byte)
}

func (bp *BytesPool) Put(v *[]byte) {
	sl := *v
	if len(sl) != bp.defaultSize {
		*v = sl[:bp.defaultSize]
	}
	bp.poll.Put(v)
}
