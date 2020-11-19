// source https://github.com/savsgio/gotils/blob/master/conv.go

package internal

import (
	"reflect"
	"sync"
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

type Buf struct {
	Data []byte
}

type _BufferPool struct {
	sync.Pool
	maxSize int
}

func (pool *_BufferPool) Get() *Buf {
	return pool.Pool.Get().(*Buf)
}

func (pool *_BufferPool) Put(buf *Buf) {
	if pool.maxSize > 0 && cap(buf.Data) > pool.maxSize {
		buf.Data = nil
	} else {
		buf.Data = buf.Data[:0]
	}
	pool.Pool.Put(buf)
}

func NewBufferPoll(maxSize int) *_BufferPool {
	return &_BufferPool{
		Pool:    sync.Pool{New: func() interface{} { return &Buf{} }},
		maxSize: maxSize,
	}
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
		if spaceMap[v[left]] != 1 {
			break
		}
	}
	for ; right >= left; right-- {
		if spaceMap[v[right]] != 1 {
			break
		}
	}
	return v[left : right+1]
}
