package sha

import (
	"bufio"
	"context"
	"errors"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"io"
	"mime"
	"net"
	"time"
)

type RequestCtx struct {
	ctx        context.Context
	cancelFunc func()

	readBuf []byte
	conn    net.Conn
	r       *bufio.Reader
	w       *bufio.Writer

	Request  Request
	Response Response

	// user data
	ud userData

	// time
	connTime time.Time

	isTLS bool

	hijacked bool

	// err
	err interface{}

	//
	keepByUser bool
}

func (ctx *RequestCtx) Keep() {
	ctx.keepByUser = true
}

func (ctx *RequestCtx) ReturnTo(pool *RequestCtxPool) {
	ctx.keepByUser = false
	pool.Release(ctx)
}

// context.Context
func (ctx *RequestCtx) SetParentContext(pctx context.Context) {
	ctx.ctx = pctx
}

func (ctx *RequestCtx) Deadline() (deadline time.Time, ok bool) { return ctx.ctx.Deadline() }

func (ctx *RequestCtx) Done() <-chan struct{} { return ctx.ctx.Done() }

func (ctx *RequestCtx) Err() error { return ctx.ctx.Err() }

func (ctx *RequestCtx) Value(key interface{}) interface{} { return ctx.ctx.Value(key) }

func (ctx *RequestCtx) SetError(v interface{}) { ctx.err = v }

type _RCtxKeyT int

const (
	_RCtxKey = _RCtxKeyT(iota)
)

func (ctx *RequestCtx) Context() context.Context {
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
func (ctx *RequestCtx) Close() { ctx.Response.Header().Set(HeaderConnection, []byte("close")) }

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
	v, ok := ctx.Request.Header().Get(HeaderConnection)
	if !ok {
		return ""
	}
	if string(inPlaceLowercase(v)) != lowerUpgradeHeader {
		return ""
	}
	v, ok = ctx.Request.Header().Get(HeaderUpgrade)
	if !ok {
		return ""
	}
	return utils.S(inPlaceLowercase(v))
}

func (ctx *RequestCtx) prepareForNextRequest() {
	ctx.Request.Reset()
	ctx.Response.reset()
	ctx.ud.Reset()
}

func (ctx *RequestCtx) resetConn() {
	ctx.ctx = nil
	ctx.cancelFunc = nil
	ctx.conn = nil
	ctx.connTime = time.Time{}
	ctx.r.Reset(nil)
	ctx.w.Reset(nil)
	ctx.Response.header.fromOutSide = false
	ctx.Request.header.fromOutSide = false
}

func (ctx *RequestCtx) Reset() {
	if ctx.ctx == nil {
		return
	}
	ctx.prepareForNextRequest()
	ctx.resetConn()
}

func (ctx *RequestCtx) SetConnection(conn net.Conn) {
	ctx.r.Reset(conn)
	ctx.w.Reset(conn)
	ctx.connTime = time.Now()
	ctx.conn = conn
}

func (ctx *RequestCtx) Write(p []byte) (int, error) {
	return ctx.Response.Write(p)
}

func (ctx *RequestCtx) WriteString(s string) (int, error) {
	return ctx.Write(utils.B(s))
}

func (ctx *RequestCtx) WriteJSON(v interface{}) {
	ctx.Response.Header().SetContentType(MIMEJson)

	encoder := jsonx.NewEncoder(ctx)
	err := encoder.Encode(v)
	if err != nil {
		panic(err)
	}
}

func (ctx *RequestCtx) WriteHTML(v []byte) {
	ctx.Response.Header().SetContentType(MIMEHtml)
	_, e := ctx.Write(v)
	if e != nil {
		panic(e)
	}
}

func (ctx *RequestCtx) WriteFile(f io.Reader, ext string) {
	ctx.Response.Header().SetContentType(mime.TypeByExtension(ext))

	buf := ctx.readBuf
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

type Template interface {
	Execute(ctx context.Context, v interface{}) error
}

func (ctx *RequestCtx) WriteTemplate(t Template, data interface{}) error {
	ctx.Response.Header().SetContentType(MIMEHtml)
	return t.Execute(ctx, data)
}
