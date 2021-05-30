package sha

import (
	"bytes"
	"sync"
	"time"
)

var UniqueIDGenerator interface {
	Size() int
	Generate([]byte)
}

type _HTTPPocket struct {
	fl1    []byte
	fl2    []byte
	fl3    []byte
	header Header
	body   *bytes.Buffer
	time   int64
	guid   []byte
}

var bodyBufPool = sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}

func (p *_HTTPPocket) GUID() []byte {
	if UniqueIDGenerator == nil {
		return nil
	}

	if len(p.guid) < 1 {
		if cap(p.guid) < UniqueIDGenerator.Size() {
			p.guid = make([]byte, UniqueIDGenerator.Size())
		}
		p.guid = p.guid[:UniqueIDGenerator.Size()]
		UniqueIDGenerator.Generate(p.guid)
	}
	return p.guid
}

func (p *_HTTPPocket) ResetGUID() { p.guid = p.guid[:0] }

func (p *_HTTPPocket) reset() {
	p.fl1 = p.fl1[:0]
	p.fl2 = p.fl2[:0]
	p.fl3 = p.fl3[:0]
	p.time = 0
	p.guid = p.guid[:0]

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
