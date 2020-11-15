package suna

import (
	"context"
	"github.com/zzztttkkk/suna/internal"
	"net"
	"net/http"
	"sync"
	"time"
)

type Request struct {
	Header Header
	Method []byte
	Path   []byte
	Query  UrlencodedForm
	Params internal.Kvs

	rawPath []byte
	version []byte
}

type Response struct {
	statusCode   int
	Header       Header
	buf          []byte
	compressW    WriteFlusher
	compressFunc func(response *Response) WriteFlusher

	headerWritten bool
}

func (res *Response) Write(p []byte) (int, error) {
	if res.compressW != nil {
		return res.compressW.Write(p)
	}
	res.buf = append(res.buf, p...)
	return len(p), nil
}

func (res *Response) SetStatusCode(v int) {
	res.statusCode = v
}

func (res *Response) ResetBodyBuffer() {
	if res.compressW != nil {
		_ = res.compressW.Flush()
		res.compressW = res.compressFunc(res)
	}
	res.buf = res.buf[:0]
}

func (res *Response) Reset() {
	res.statusCode = 0
	res.Header.Reset()
	res.headerWritten = false

	res.ResetBodyBuffer()
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
	conn    net.Conn
	connCtx context.Context

	// parse
	status       int
	fStatus      int // first line status
	fLSize       int // first line size
	hSize        int // header lines size
	parseBuf     []byte
	cHKey        []byte // current header key
	cHKeyDoUpper bool   // prev byte is '-' or first byte
	kvSep        bool   // `:`
	bodyRemain   int
	bodySize     int
}

func (ctx *RequestCtx) RemoteAddr() net.Addr {
	return ctx.conn.RemoteAddr()
}

var connectionHeader = []byte("Connection")
var upgradeHeader = []byte("Upgrade")

// Upgrade return false, if not an upgrade request
func (ctx *RequestCtx) Upgrade() (Protocol, bool) {
	v, ok := ctx.Request.Header.Get(connectionHeader)
	if !ok {
		return nil, false
	}
	if string(v) != string(upgradeHeader) {
		return nil, false
	}
	v, ok = ctx.Request.Header.Get(upgradeHeader)
	if !ok || len(v) < 1 {
		return nil, false
	}
	v = inplaceLowercase(v)
	protocol := ctx.protocol.SubProtocols[internal.S(v)]
	if protocol == nil {
		ctx.WriteError(StdHttpErrors[http.StatusNotFound])
		return nil, false
	}

	if !protocol.Handshake(ctx) {
		return protocol, false
	}
	return protocol, true
}

func (ctx *RequestCtx) reset() {
	ctx.Context = nil

	ctx.Request.Header.Reset()
	ctx.Request.Method = ctx.Request.Method[:0]
	ctx.Request.Path = ctx.Request.Path[0:]
	ctx.Request.Query.Reset()
	ctx.Request.Params.Reset()
	ctx.Request.rawPath = ctx.Request.rawPath[:0]
	ctx.Request.version = ctx.Request.version[:0]

	ctx.Response.Reset()
	ctx.Response.compressW = nil
	ctx.Response.compressFunc = nil

	ctx.status = 0
	ctx.fStatus = 0
	ctx.fLSize = 0
	ctx.hSize = 0
	ctx.parseBuf = ctx.parseBuf[:0]
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
	if cap(ctx.parseBuf) > MaxRequestCtxPooledBufferSize {
		ctx.parseBuf = nil
	}
	// response body buffer
	if cap(ctx.Response.buf) > MaxRequestCtxPooledBufferSize {
		ctx.Response.buf = nil
	}

	ctx.makeRequestCtx = nil
	ctx.conn = nil
	ctx.connCtx = nil

	ctxPool.Put(ctx)
}

type RequestHandler interface {
	Handle(ctx *RequestCtx)
}

type RequestHandlerFunc func(ctx *RequestCtx)

func (fn RequestHandlerFunc) Handle(ctx *RequestCtx) {
	fn(ctx)
}
