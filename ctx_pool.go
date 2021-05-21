package sha

import (
	"bufio"
	"context"
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

var defaultRCtxPool *RequestCtxPool

func init() {
	defaultRCtxPool = NewRequestCtxPool(nil)
}

func AcquireRequestCtx(ctx context.Context) *RequestCtx {
	rctx := defaultRCtxPool.Acquire()
	rctx.ctx = ctx
	return rctx
}

func ReleaseRequestCtx(ctx *RequestCtx) {
	ctx.Reset()
	defaultRCtxPool.Put(ctx)
}

func (p *RequestCtxPool) Acquire() *RequestCtx {
	ctx := p.Get().(*RequestCtx)
	if ctx.readBuf == nil {
		ctx.readBuf = make([]byte, p.readSize)
	}
	if ctx.r == nil {
		ctx.r = bufio.NewReaderSize(nil, p.readSize)
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

	if p.opt.BufferPoolSizeLimit > 0 {
		if ctx.Request.body != nil && ctx.Request.body.Cap() > p.opt.BufferPoolSizeLimit {
			ctx.Request.body = nil
		}
		if ctx.Response.body != nil && ctx.Response.body.Cap() > p.opt.BufferPoolSizeLimit {
			ctx.Response.body = nil
		}
	}

	if unprepared {
		ctx.prepareForNextRequest()
	}
	ctx.resetConn()
	p.Put(ctx)
}
