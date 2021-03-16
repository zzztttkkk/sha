package sha

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zzztttkkk/sha/utils"
	"html/template"
	"io"
	"mime"
	"net"
	"net/http"
	"sync"
	"time"
)

type RequestCtx struct {
	ctx      context.Context
	Request  Request
	Response Response
	ud       userData

	// time
	connTime time.Time
	reqTime  time.Time

	// writer
	isTLS    bool
	conn     net.Conn
	hijacked bool

	// parser
	status           int
	firstLineStatus  int
	firstLineSize    int // first line size
	headersSize      int // header lines size
	buf              []byte
	currentHeaderKey []byte // current header key
	cHKeyDoUpper     bool   // prev byte is '-' or first byte
	headerKVSepRead  bool   // `:`
	bodyRemain       int
	bodySize         int

	// hook
	onReset []func(ctx *RequestCtx)

	// err
	err interface{}
}

// context.Context
func (ctx *RequestCtx) Deadline() (deadline time.Time, ok bool) { return ctx.ctx.Deadline() }

func (ctx *RequestCtx) Done() <-chan struct{} { return ctx.ctx.Done() }

func (ctx *RequestCtx) Err() error { return ctx.ctx.Err() }

func (ctx *RequestCtx) Value(key interface{}) interface{} { return ctx.ctx.Value(key) }

func (ctx *RequestCtx) Error(v interface{}) { ctx.err = v }

type _RCtxKeyT int

const (
	_RCtxKey = _RCtxKeyT(iota)
)

func Wrap(ctx *RequestCtx) context.Context {
	return context.WithValue(ctx, _RCtxKey, ctx)
}

func Unwrap(ctx context.Context) *RequestCtx {
	c, ok := ctx.(*RequestCtx)
	if ok {
		return c
	}
	v := ctx.Value(_RCtxKey)
	if v != nil {
		return v.(*RequestCtx)
	}
	return nil
}

// custom data
func (ctx *RequestCtx) SetCustomData(key string, value interface{}) { ctx.ud.Set(key, value) }

func (ctx *RequestCtx) GetCustomData(key string) interface{} { return ctx.ud.Get(key) }

func (ctx *RequestCtx) VisitCustomData(fn func(key []byte, v interface{}) bool) { ctx.ud.Visit(fn) }

// http connection
func (ctx *RequestCtx) Close() { ctx.Response.Header.Set(HeaderConnection, []byte("close")) }

func (ctx *RequestCtx) IsTLS() bool { return ctx.isTLS }

func (ctx *RequestCtx) RemoteAddr() net.Addr { return ctx.conn.RemoteAddr() }

var ErrRequestHijacked = errors.New("sha: request is already hijacked")

func (ctx *RequestCtx) Hijack() net.Conn {
	if ctx.hijacked {
		panic(ErrRequestHijacked)
	}
	ctx.hijacked = true
	return ctx.conn
}

const lowerUpgradeHeader = "upgrade"

func (ctx *RequestCtx) UpgradeProtocol() string {
	if string(ctx.Request.Method) != http.MethodGet {
		return ""
	}
	v, ok := ctx.Request.Header.Get(HeaderConnection)
	if !ok {
		return ""
	}
	if string(inPlaceLowercase(v)) != lowerUpgradeHeader {
		return ""
	}
	v, ok = ctx.Request.Header.Get(HeaderUpgrade)
	if !ok {
		return ""
	}
	return utils.S(inPlaceLowercase(v))
}

func (ctx *RequestCtx) OnReset(fn func(ctx *RequestCtx)) { ctx.onReset = append(ctx.onReset, fn) }

func (ctx *RequestCtx) Reset() {
	if ctx.ctx == nil {
		return
	}

	if len(ctx.onReset) > 0 {
		for _, fn := range ctx.onReset {
			fn(ctx)
		}
		ctx.onReset = ctx.onReset[:0]
	}

	ctx.ctx = nil
	ctx.Request.Reset()
	ctx.Response.reset()
	ctx.ud.Reset()

	ctx.status = 0
	ctx.firstLineStatus = 0
	ctx.firstLineSize = 0
	ctx.headersSize = 0
	ctx.buf = ctx.buf[:0]
	ctx.currentHeaderKey = ctx.currentHeaderKey[:0]
	ctx.cHKeyDoUpper = false
	ctx.headerKVSepRead = false
	ctx.bodySize = -1
	ctx.bodyRemain = -1
}

var ctxPool = sync.Pool{New: func() interface{} { return &RequestCtx{} }}

func acquireRequestCtx() *RequestCtx { return ctxPool.Get().(*RequestCtx) }

func ReleaseRequestCtx(ctx *RequestCtx) {
	ctx.Reset()
	ctx.Response.freeWriter()
	ctx.conn = nil
	ctx.isTLS = false
	ctxPool.Put(ctx)
}

type RequestHandler interface {
	Handle(ctx *RequestCtx)
}

type RequestHandlerFunc func(ctx *RequestCtx)

func (fn RequestHandlerFunc) Handle(ctx *RequestCtx) {
	fn(ctx)
}

func (ctx *RequestCtx) Write(p []byte) (int, error) {
	return ctx.Response.Write(p)
}

func (ctx *RequestCtx) WriteString(s string) (int, error) {
	return ctx.Write(utils.B(s))
}

func (ctx *RequestCtx) WriteJSON(v interface{}) {
	ctx.Response.Header.SetContentType(MIMEJson)

	encoder := json.NewEncoder(ctx)
	err := encoder.Encode(v)
	if err != nil {
		panic(err)
	}
}

func (ctx *RequestCtx) WriteHTML(v []byte) {
	ctx.Response.Header.SetContentType(MIMEHtml)
	_, e := ctx.Write(v)
	if e != nil {
		panic(e)
	}
}

func (ctx *RequestCtx) WriteFile(f io.Reader, ext string) {
	ctx.Response.Header.SetContentType(mime.TypeByExtension(ext))

	buf := make([]byte, 512, 512)
	for {
		l, e := f.Read(buf)
		if e != nil {
			panic(e)
		}
		_, e = ctx.Write(buf[:l])
		if e != nil {
			panic(e)
		}
	}
}

func (ctx *RequestCtx) WriteTemplate(t *template.Template, data interface{}) {
	ctx.Response.Header.SetContentType(MIMEHtml)
	e := t.Execute(ctx, data)
	if e != nil {
		panic(e)
	}
}

func (ctx *RequestCtx) SetStatus(status int) { ctx.Response.statusCode = status }

func (ctx *RequestCtx) GetStatus() int { return ctx.Response.statusCode }
