package suna

import (
	"context"
	"github.com/zzztttkkk/suna/internal"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Request struct {
	Header Header
	Method []byte
	Path   []byte
	Params internal.Kvs

	cookies internal.Kvs
	query   Form
	body    Form
	files   FormFiles

	queryStatus   int // >2: `?` index; 1: parsed; 0 empty
	bodyStatus    int // 0: unparsed; 1: unsupported content type; 2: parsed
	cookieParsed  bool
	rawPath       []byte
	version       []byte
	bodyBufferPtr *[]byte

	// websocket
	wsSubP       SubWebSocketProtocol
	wsDoCompress bool
}

func (req *Request) Reset() {
	req.Header.Reset()
	req.Method = req.Method[:0]
	req.Path = req.Path[:0]
	req.Params.Reset()

	req.cookies.Reset()
	req.query.Reset()
	req.body.Reset()
	req.files = nil
	req.files = nil
	req.cookieParsed = false
	req.queryStatus = 0
	req.bodyStatus = 0
	req.rawPath = req.rawPath[:0]
	req.version = req.version[:0]
	req.bodyBufferPtr = nil
	req.wsSubP = nil
	req.wsDoCompress = false
}

func (req *Request) Cookie(key []byte) ([]byte, bool) {
	if !req.cookieParsed {
		v, ok := req.Header.Get(internal.B(HeaderCookie))
		if ok {
			var key []byte
			var buf []byte

			for _, b := range v {
				switch b {
				case '=':
					key = append(key, buf...)
					buf = buf[:0]
				case ';':
					req.cookies.Set(decodeURI(key), decodeURI(buf))
					key = key[:0]
					buf = buf[:0]
				case ' ':
					continue
				default:
					buf = append(buf, b)
				}
			}
			req.cookies.Set(decodeURI(key), decodeURI(buf))
		}
		req.cookieParsed = true
	}
	return req.cookies.Get(key)
}

func (ctx *RequestCtx) Cookie(key []byte) ([]byte, bool) {
	return ctx.Request.Cookie(key)
}

type Response struct {
	statusCode int
	Header     Header

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

type _SameSiteVal string

const (
	CookeSameSizeDefault = _SameSiteVal("")
	CookieSameSiteLax    = _SameSiteVal("lax")
	CookieSameSiteStrict = _SameSiteVal("strict")
	CookieSameSizeNone   = _SameSiteVal("none")
)

type CookieOptions struct {
	Domain   string
	Path     string
	MaxAge   int64
	Expires  time.Time
	Secure   bool
	HttpOnly bool
	SameSite _SameSiteVal
}

func (res *Response) SetCookie(k, v string, options CookieOptions) {
	item := res.Header.Append(internal.B(HeaderSetCookie), nil)

	item.Val = append(item.Val, internal.B(k)...)
	item.Val = append(item.Val, '=')
	item.Val = append(item.Val, internal.B(v)...)
	item.Val = append(item.Val, ';')
	item.Val = append(item.Val, ' ')

	if len(options.Domain) > 0 {
		item.Val = append(item.Val, 'D', 'o', 'm', 'a', 'i', 'n', '=')
		item.Val = append(item.Val, internal.B(options.Domain)...)
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	}

	if len(options.Path) > 0 {
		item.Val = append(item.Val, 'P', 'a', 't', 'h', '=')
		item.Val = append(item.Val, internal.B(options.Path)...)
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	}

	if !options.Expires.IsZero() {
		item.Val = append(item.Val, 'E', 'x', 'p', 'i', 'r', 'e', 's', '=')
		item.Val = append(item.Val, internal.B(options.Expires.Format(time.RFC1123))...)
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	} else {
		item.Val = append(item.Val, 'M', 'a', 'x', '-', 'A', 'g', 'e', '=')
		item.Val = append(item.Val, internal.B(strconv.FormatInt(options.MaxAge, 10))...)
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	}

	if options.Secure {
		item.Val = append(item.Val, 'S', 'e', 'c', 'u', 'r', 'e')
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	}

	if options.HttpOnly {
		item.Val = append(item.Val, 'H', 't', 't', 'p', 'o', 'n', 'l', 'y')
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	}

	if len(options.SameSite) > 0 {
		item.Val = append(item.Val, 'S', 'a', 'm', 'e', 's', 'i', 't', 'e', '=')
		item.Val = append(item.Val, internal.B(string(options.SameSite))...)
		item.Val = append(item.Val, ';')
		item.Val = append(item.Val, ' ')
	}
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
	hijacked     bool
}

func (ctx *RequestCtx) RemoteAddr() net.Addr {
	return ctx.conn.RemoteAddr()
}

const lowerUpgradeHeader = "upgrade"

func (ctx *RequestCtx) UpgradeProtocol() string {
	if string(ctx.Request.Method) != http.MethodGet {
		return ""
	}
	v, ok := ctx.Request.Header.Get(internal.B(HeaderConnection))
	if !ok {
		return ""
	}
	if string(inplaceLowercase(v)) != lowerUpgradeHeader {
		return ""
	}
	v, ok = ctx.Request.Header.Get(internal.B(HeaderUpgrade))
	if !ok {
		return ""
	}
	return internal.S(inplaceLowercase(v))
}

func (ctx *RequestCtx) Reset() {
	if ctx.Context == nil {
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
}

var ctxPool = sync.Pool{New: func() interface{} { return &RequestCtx{} }}

func AcquireRequestCtx() *RequestCtx {
	v := ctxPool.Get().(*RequestCtx)
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
