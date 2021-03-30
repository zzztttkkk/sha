package sha

import (
	"bytes"
	"sync"
	"time"
)

type _HTTPPocket struct {
	fl1    []byte
	fl2    []byte
	fl3    []byte
	header Header
	body   *bytes.Buffer
	time   int64
}

var bodyBufPool = sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}

func (p *_HTTPPocket) reset() {
	p.fl1 = p.fl1[:0]
	p.fl2 = p.fl2[:0]
	p.fl3 = p.fl3[:0]
	p.time = 0

	p.header.Reset()

	if p.body != nil {
		p.body.Reset()
		bodyBufPool.Put(p.body)
		p.body = nil
	}
}

func (p *_HTTPPocket) Write(v []byte) (int, error) {
	if p.body == nil {
		p.body = bodyBufPool.Get().(*bytes.Buffer)
	}
	return p.body.Write(v)
}

func (p *_HTTPPocket) Header() *Header { return &p.header }

func (p *_HTTPPocket) UnixNano() int64 { return p.time }

func (p *_HTTPPocket) setTime() { p.time = time.Now().UnixNano() }
