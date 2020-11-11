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
