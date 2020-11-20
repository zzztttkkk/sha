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
	Params internal.Kvs

	query Form
	body  Form
	files FormFiles

	queryStatus   int // >2: `?` index; 1: parsed; 0 empty
	bodyStatus    int // 0: unparsed; 1: unsupported content type; 2: parsed
	rawPath       []byte
	version       []byte
	bodyBufferPtr *[]byte
}

func (req *Request) Reset() {
	req.Header.Reset()
	req.Method = req.Method[:0]
	req.Path = req.Path[:0]
	req.Params.Reset()

	req.query.Reset()
	req.body.Reset()
	req.files = nil
	req.files = nil
	req.queryStatus = 0
	req.bodyStatus = 0
	req.rawPath = req.rawPath[:0]
	req.version = req.version[:0]
	req.bodyBufferPtr = nil
}

type Response struct {
	statusCode        int
	Header            Header
	buf               []byte
	compressWriter    WriteFlusher
	newCompressWriter func(response *Response) WriteFlusher

	headerWritten bool
}

func (res *Response) Write(p []byte) (int, error) {
	if res.compressWriter != nil {
		return res.compressWriter.Write(p)
	}
	res.buf = append(res.buf, p...)
	return len(p), nil
}

func (res *Response) SetStatusCode(v int) {
	res.statusCode = v
}

func (res *Response) ResetBodyBuffer() {
	if res.compressWriter != nil {
		_ = res.compressWriter.Flush()
		res.compressWriter = res.newCompressWriter(res)
	}
	res.buf = res.buf[:0]
}

//Reset after executing `Response.Reset`, we still need to keep the compression
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
	buf          []byte
	cHKey        []byte // current header key
	cHKeyDoUpper bool   // prev byte is '-' or first byte
	kvSep        bool   // `:`
	bodyRemain   int
	bodySize     int
	reset        bool
}

func (ctx *RequestCtx) RemoteAddr() net.Addr {
	return ctx.conn.RemoteAddr()
}

var headerConnection = []byte("Connection")
var upgradeHeader = []byte("Upgrade")

// Upgrade return false, if not an upgrade request
func (ctx *RequestCtx) Upgrade() (Protocol, bool) {
	v, ok := ctx.Request.Header.Get(headerConnection)
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
		ctx.WriteStatus(http.StatusNotFound)
		return nil, false
	}

	if !protocol.Handshake(ctx) {
		return protocol, false
	}
	return protocol, true
}

func (ctx *RequestCtx) Reset() {
	if ctx.reset {
		return
	}

	ctx.Context = nil

	ctx.Request.Reset()

	ctx.Response.Reset()
	ctx.Response.compressWriter = nil
	ctx.Response.newCompressWriter = nil

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

	ctx.reset = true
}

var ctxPool = sync.Pool{New: func() interface{} { return &RequestCtx{} }}

func AcquireRequestCtx() *RequestCtx {
	v := ctxPool.Get().(*RequestCtx)
	v.reset = false
	return v
}

var MaxRequestCtxPooledBufferSize = 1024 * 4

func ReleaseRequestCtx(ctx *RequestCtx) {
	ctx.Reset()
	// request body buffer
	if cap(ctx.buf) > MaxRequestCtxPooledBufferSize {
		ctx.buf = nil
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
