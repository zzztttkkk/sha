package utils

import (
	"sync"
)

type Buf struct {
	Data []byte
}

func (b *Buf) Write(p []byte) (int, error) {
	b.Data = append(b.Data, p...)
	return len(p), nil
}

func (b *Buf) WriteString(v string) {
	b.Data = append(b.Data, v...)
}

func (b *Buf) WriteByte(v byte) error {
	b.Data = append(b.Data, v)
	return nil
}

func (b *Buf) Reset() {
	b.Data = b.Data[:0]
}

type BufferPool struct {
	sync.Pool
	maxSize int
}

func (pool *BufferPool) Get() *Buf {
	return pool.Pool.Get().(*Buf)
}

func (pool *BufferPool) Put(buf *Buf) {
	if pool.maxSize > 0 && cap(buf.Data) > pool.maxSize {
		buf.Data = nil
	} else {
		buf.Data = buf.Data[:0]
	}
	pool.Pool.Put(buf)
}

func NewBufferPoll(maxSize int) *BufferPool {
	return &BufferPool{
		Pool:    sync.Pool{New: func() interface{} { return &Buf{Data: nil} }},
		maxSize: maxSize,
	}
}

type FixedSizeBufferPool struct {
	BufferPool
	defaultSize int
}

func (pool *FixedSizeBufferPool) Put(buf *Buf) {
	if pool.maxSize > 0 && cap(buf.Data) > pool.maxSize {
		buf.Data = make([]byte, pool.defaultSize)
	}
	pool.Pool.Put(buf)
}

func NewFixedSizeBufferPoll(defaultSize, maxSize int) *FixedSizeBufferPool {
	pool := &FixedSizeBufferPool{defaultSize: defaultSize}
	pool.New = func() interface{} { return &Buf{Data: make([]byte, defaultSize)} }
	pool.maxSize = maxSize
	return pool
}
