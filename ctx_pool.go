package sha

import (
	"bufio"
	"context"
	"github.com/zzztttkkk/sha/utils"
	"sync"
)

type RequestCtxPool struct {
	sync.Pool
	opt HTTPOptions
}

func NewRequestCtxPool(opt *HTTPOptions) *RequestCtxPool {
	pool := &RequestCtxPool{Pool: sync.Pool{New: func() interface{} { return &RequestCtx{} }}}
	if opt == nil {
		pool.opt = defaultHTTPOption
	} else {
		pool.opt = *opt
		utils.Merge(&pool.opt, defaultHTTPOption)
	}
	return pool
}

var defaultRCtxPool = NewRequestCtxPool(nil)

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
		ctx.readBuf = make([]byte, p.opt.ReadBufferSize)
	}
	if ctx.r == nil {
		ctx.r = bufio.NewReaderSize(nil, p.opt.ReadBufferSize)
	}
	if ctx.w == nil {
		ctx.w = bufio.NewWriterSize(nil, p.opt.ReadBufferSize)
	}
	return ctx
}

func (p *RequestCtxPool) Release(ctx *RequestCtx) { p.release(ctx, true) }

func (p *RequestCtxPool) release(ctx *RequestCtx, unprepared bool) {
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
