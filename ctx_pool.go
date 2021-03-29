package sha

import (
	"bufio"
	"sync"
)

type RequestCtxPool struct {
	sync.Pool
	readSize  int
	writeSize int
	opt       *HTTPOption
}

func NewRequestCtxPool(opt *HTTPOption) *RequestCtxPool {
	if opt == nil {
		opt = &defaultHTTPOption
	}
	var eO HTTPOption
	eO = *opt
	return &RequestCtxPool{
		Pool:      sync.Pool{New: func() interface{} { return &RequestCtx{} }},
		readSize:  eO.ReadBufferSize,
		writeSize: eO.SendBufferSize,
		opt:       &eO,
	}
}

var _defaultRCtxPool *RequestCtxPool
var __defaultRCtxPoolOnce sync.Once

func DefaultRequestCtxPool() *RequestCtxPool {
	__defaultRCtxPoolOnce.Do(func() { _defaultRCtxPool = NewRequestCtxPool(nil) })
	return _defaultRCtxPool
}

func (p *RequestCtxPool) Acquire() *RequestCtx {
	ctx := p.Get().(*RequestCtx)
	if ctx.readBuf == nil {
		ctx.readBuf = make([]byte, p.readSize)
	}
	if ctx.r == nil {
		ctx.r = bufio.NewReaderSize(nil, p.readSize*2)
	}
	if ctx.w == nil {
		ctx.w = bufio.NewWriterSize(nil, p.writeSize)
	}
	return ctx
}

func (p *RequestCtxPool) Release(ctx *RequestCtx) {
	if ctx.keepByUser {
		return
	}

	if p.opt.MaxBodyBufferSize > 0 {
		if ctx.Request.body != nil && ctx.Request.body.Cap() > p.opt.MaxBodyBufferSize {
			ctx.Request.body = nil
		}
		if ctx.Response.body != nil && ctx.Response.body.Cap() > p.opt.MaxBodyBufferSize {
			ctx.Response.body = nil
		}
	}
	ctx.Reset()
	p.Put(ctx)
}

func (p *RequestCtxPool) NewHTTPSession(address string, env *Environment) *HTTPSession {
	return newHTTPSession(address, p, env)
}

func (p *RequestCtxPool) NewHTTPSSession(address string, env *Environment) *HTTPSession {
	return newHTTPSSession(address, p, env)
}

func (p *RequestCtxPool) NewHTTPServerProtocol() {

}
