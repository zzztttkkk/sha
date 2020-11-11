package suna

import (
	"bytes"
	"context"
	"errors"
	"github.com/zzztttkkk/suna/internal"
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
	statusCode int
	Header     Header
}

type RequestCtx struct {
	context.Context
	Request Request
	Response

	ctxFun func() context.Context

	// time
	connTime time.Time
	reqTime  time.Time

	// writer
	noBuffer      bool
	conn          net.Conn
	wBuf          []byte
	headerWritten bool

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

func (ctx *RequestCtx) writeHeader() {

}

func (ctx *RequestCtx) Write(p []byte) (int, error) {
	if ctx.noBuffer {
		if !ctx.headerWritten {
			ctx.writeHeader()
			ctx.headerWritten = true
		}
		return ctx.conn.Write(p)
	}
	ctx.wBuf = append(ctx.wBuf, p...)
	return len(p), nil
}

func (ctx *RequestCtx) RemoteAddr() net.Addr {
	return ctx.conn.RemoteAddr()
}

var responseBufferPool = internal.NewBufferPoll(1024)

func (ctx *RequestCtx) sendResponseBuffer() {
	if ctx.noBuffer {
		return
	}

	buf := responseBufferPool.Get()
	defer responseBufferPool.Put(buf)
}

func (ctx *RequestCtx) prepareForNext() {
	ctx.sendResponseBuffer()

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

	ctx.Response.statusCode = 0
	ctx.Response.Header.Reset()

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

var spaceMap []byte

func init() {
	for i := 0; i < 256; i++ {
		switch i {
		case '\t', '\n', '\v', '\f', '\r', ' ':
			spaceMap = append(spaceMap, 1)
		default:
			spaceMap = append(spaceMap, 0)
		}
	}
}

func inplaceTrimSpace(v []byte) []byte {
	var left = 0
	var right = len(v) - 1
	for ; left < right; left++ {
		if spaceMap[v[left]] != 1 {
			break
		}
	}
	for ; right > left; right-- {
		if spaceMap[v[right]] != 1 {
			break
		}
	}
	return v[left : right+1]
}

func (ctx *RequestCtx) onRequestHeaderLine() {
	key := inplaceTrimSpace(ctx.cHKey)
	val := inplaceTrimSpace(ctx.buf)
	ctx.Request.Header.Append(key, val)
}

var ErrBadConnection = errors.New("bad connection")
var ErrUrlTooLong = errors.New("url too long")                                     // 414
var ErrRequestHeaderFieldsTooLarge = errors.New("request header fields too large") // 431
var ErrPayloadTooLarge = errors.New("payload too large")                           // 413
var httpVersion = []byte("http/")

func (ctx *RequestCtx) initRequest() {
	ctx.Context = ctx.ctxFun()
	ctx.reqTime = time.Now()
	ctx.bodySize = ctx.Request.Header.ContentLength()
	ctx.bodyRemain = ctx.bodySize

	ctx.Request.URI.init(ctx.Request.rawPath, true)
}

func (ctx *RequestCtx) feedReqData(data []byte, offset, end int, protocol *HttpProtocol) (int, error) {
	var v byte

	switch ctx.status {
	case 0: // parse first line
		for ; offset < end; {
			v = data[offset]
			offset++
			ctx.fLSize++
			if ctx.fLSize > protocol.MaxFirstLintSize {
				return 1, ErrUrlTooLong
			}
			if v > 126 || v < 10 {
				return 45, ErrBadConnection
			}
			if v == '\n' {
				ctx.status++
				ctx.buf = ctx.buf[:0]
				if len(ctx.Request.rawPath) < 1 || ctx.Request.rawPath[0] != '/' { // empty path
					return 2, ErrBadConnection
				}
				if !bytes.HasPrefix(inplaceLowercase(ctx.Request.version), httpVersion) { // http version
					return 3, ErrBadConnection
				}
				return offset, nil
			}

			switch v {
			case '\r':
			case ' ':
				ctx.fStatus += 1
			default:
				switch ctx.fStatus {
				case 0:
					ctx.Request.Method = append(ctx.Request.Method, v)
				case 1:
					ctx.Request.rawPath = append(ctx.Request.rawPath, v)
				case 2:
					ctx.Request.version = append(ctx.Request.version, v)
				default:
					return 4, ErrBadConnection
				}
			}

			if v != '\r' {
				ctx.buf = append(ctx.buf, v)
			}
		}
	case 1: // parse header line
		for ; offset < end; {
			v = data[offset]
			offset++
			ctx.hSize++
			if ctx.hSize > protocol.MaxHeadersSize {
				return 5, ErrRequestHeaderFieldsTooLarge
			}
			if v > 126 || v < 10 {
				return 45, ErrBadConnection
			}

			if v == '\n' {
				if len(ctx.cHKey) < 1 { // all header data read done
					ctx.status++
					return offset, nil
				}
				ctx.onRequestHeaderLine()
				ctx.cHKey = ctx.cHKey[:0]
				ctx.buf = ctx.buf[:0]
				return offset, nil
			}

			if v == '\r' {
				ctx.kvSep = false
				ctx.cHKeyDoUpper = true
				continue
			}

			if !ctx.kvSep {
				if v == ':' {
					ctx.kvSep = true
				} else {
					if ctx.cHKeyDoUpper {
						ctx.cHKeyDoUpper = false
						v = toUpperTable[v]
					}
					ctx.cHKey = append(ctx.cHKey, v)
					if v == '-' {
						ctx.cHKeyDoUpper = true
					}
				}
			} else {
				ctx.buf = append(ctx.buf, v)
			}
		}
	case 2:
		if ctx.Context == nil {
			ctx.initRequest()
			if ctx.bodySize > protocol.MaxBodySize {
				return 6, ErrPayloadTooLarge
			}
		}

		size := end - offset
		if size > ctx.bodyRemain {
			return 7, ErrBadConnection
		}
		ctx.buf = append(ctx.buf, data[offset:end]...)
		ctx.bodyRemain -= size
		return end, nil
	}
	return offset, nil
}

var ctxPool = sync.Pool{New: func() interface{} { return &RequestCtx{} }}

func AcquireRequestCtx() *RequestCtx {
	return ctxPool.Get().(*RequestCtx)
}

var MaxPooledBufferSize = 1024 * 4

func ReleaseRequestCtx(ctx *RequestCtx) {
	// request body buffer
	if cap(ctx.buf) > MaxPooledBufferSize {
		ctx.buf = nil
	} else {
		ctx.buf = ctx.buf[:0]
	}

	// response body buffer
	if cap(ctx.wBuf) > MaxPooledBufferSize {
		ctx.wBuf = nil
	} else {
		ctx.wBuf = ctx.wBuf[:0]
	}
	ctx.prepareForNext()
	ctxPool.Put(ctx)
}

type RequestHandler interface {
	Handle(ctx *RequestCtx)
}

type RequestHandlerFunc func(ctx *RequestCtx)

func (fn RequestHandlerFunc) Handle(ctx *RequestCtx) {
	fn(ctx)
}
