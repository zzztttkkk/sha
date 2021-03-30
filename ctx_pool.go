package sha

import (
	"bufio"
	"sync"
)

type RequestCtxPool struct {
	sync.Pool
	readSize  int
	writeSize int
	opt       *HTTPOptions
}

func NewRequestCtxPool(opt *HTTPOptions) *RequestCtxPool {
	if opt == nil {
		opt = &defaultHTTPOption
	}
	var eO HTTPOptions
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

func AcquireRequestCtx() *RequestCtx { return DefaultRequestCtxPool().Acquire() }

func ReleaseRequestCtx(ctx *RequestCtx) { DefaultRequestCtxPool().Put(ctx) }

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
	p.release(ctx, true)
}

func (p *RequestCtxPool) release(ctx *RequestCtx, unprepared bool) {
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

	if unprepared {
		ctx.prepareForNextRequest()
	}
	ctx.resetConn()
	p.Put(ctx)
}
