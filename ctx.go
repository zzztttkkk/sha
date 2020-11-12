package suna

import (
	"context"
	"net"
	"sync"
	"time"
)

type Request struct {
	URI     URI
	Header  Header
	Method  []byte
	rawPath []byte
	version []byte
}

type Response struct {
	statusCode    int
	Header        Header
	buf           []byte
	headerWritten bool
}

func (res *Response) Reset() {
	res.statusCode = 0
	res.Header.Reset()
	res.buf = res.buf[:0]
	res.headerWritten = false
}

type RequestCtx struct {
	context.Context
	Request  Request
	Response Response

	protocol *Http1xProtocol

	makeRequestCtx func() context.Context

	// time
	connTime time.Time
	reqTime  time.Time

	// writer
	noBuffer bool
	conn     net.Conn

	// request parse
	status       int
	fStatus      int // first line status
	fLSize       int // first line size
	hSize        int // header lines size
	buf          []byte
	cHKey        []byte // current header key
	cHKeyDoUpper bool   // prev byte is '-' or first byte
	kvSep        bool   // `:`
	bodyRemain   int
	bodySize     int
}

// unsafe
func (ctx *RequestCtx) UseResponseBuffer(v bool) {
	ctx.noBuffer = !v
}

func (ctx *RequestCtx) RemoteAddr() net.Addr {
	return ctx.conn.RemoteAddr()
}

func (ctx *RequestCtx) reset() {
	ctx.Context = nil

	ctx.Request.Header.Reset()
	ctx.Request.version = ctx.Request.version[:0]
	ctx.Request.rawPath = ctx.Request.rawPath[:0]
	ctx.Request.Method = ctx.Request.Method[:0]
	ctx.Request.URI.Scheme = ctx.Request.URI.Scheme[:0]
	ctx.Request.URI.User = ctx.Request.URI.User[:0]
	ctx.Request.URI.Password = ctx.Request.URI.Password[:0]
	ctx.Request.URI.Host = ctx.Request.URI.Host[:0]
	ctx.Request.URI.Port = 0
	ctx.Request.URI.Path = ctx.Request.URI.Path[:0]
	ctx.Request.URI.Query.Reset()
	ctx.Request.URI.Fragment = ctx.Request.URI.Fragment[:0]

	ctx.noBuffer = false
	ctx.Response.Reset()

	ctx.status = 0
	ctx.fStatus = 0
	ctx.fLSize = 0
	ctx.hSize = 0
	ctx.buf = ctx.buf[:0]
	ctx.cHKey = ctx.cHKey[:0]
	ctx.cHKeyDoUpper = false
	ctx.kvSep = false
	ctx.bodySize = -1
	ctx.bodyRemain = -1
}

var ctxPool = sync.Pool{New: func() interface{} { return &RequestCtx{} }}

func AcquireRequestCtx() *RequestCtx {
	return ctxPool.Get().(*RequestCtx)
}

var MaxRequestCtxPooledBufferSize = 1024 * 4

func ReleaseRequestCtx(ctx *RequestCtx) {
	ctx.reset()
	// request body buffer
	if cap(ctx.buf) > MaxRequestCtxPooledBufferSize {
		ctx.buf = nil
	}
	// response body buffer
	if cap(ctx.Response.buf) > MaxRequestCtxPooledBufferSize {
		ctx.Response.buf = nil
	}
	ctxPool.Put(ctx)
}

type RequestHandler interface {
	Handle(ctx *RequestCtx)
}

type RequestHandlerFunc func(ctx *RequestCtx)

func (fn RequestHandlerFunc) Handle(ctx *RequestCtx) {
	fn(ctx)
}
