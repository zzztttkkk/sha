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

// RequestCtx
// most of the fields and most of the method return values are read-only, so:
// 1, if you want to modify them, you should keep this in mind.
// 2, if you want to keep them after handling, you should copy them.
type RequestCtx struct {
	noCopy

	ctx        context.Context
	cancelFunc func()

	readBuf []byte
	conn    net.Conn
	r       *bufio.Reader
	w       *bufio.Writer

	Request  Request
	Response Response

	// user data
	UserData userData

	// time
	connTime time.Time

	isTLS bool

	hijacked bool

	// err
	err interface{}

	// session
	sessionOK bool
	session   Session
}

func (ctx *RequestCtx) TimeSpent() time.Duration {
	diff := ctx.Request.time - ctx.Response.time
	if diff >= 0 {
		return time.Duration(diff)
	}
	return time.Duration(-diff)
}

func (ctx *RequestCtx) Deadline() (deadline time.Time, ok bool) { return ctx.ctx.Deadline() }

func (ctx *RequestCtx) Done() <-chan struct{} { return ctx.ctx.Done() }

func (ctx *RequestCtx) Err() error { return ctx.ctx.Err() }

func (ctx *RequestCtx) Value(key interface{}) interface{} { return ctx.ctx.Value(key) }

func (ctx *RequestCtx) SetError(v interface{}) { ctx.err = v }

func (ctx *RequestCtx) Wrap() context.Context {
	return context.WithValue(ctx, CtxKeyRequestCtx, ctx)
}

func Unwrap(ctx context.Context) *RequestCtx {
	c, ok := ctx.(*RequestCtx)
	if ok {
		return c
	}
	v := ctx.Value(CtxKeyRequestCtx)
	if v != nil {
		return v.(*RequestCtx)
	}
	return nil
}

func (ctx *RequestCtx) Close() { ctx.Response.Header().SetString(HeaderConnection, headerValClose) }

func (ctx *RequestCtx) IsTLS() bool { return ctx.isTLS }

func (ctx *RequestCtx) RemoteAddr() net.Addr { return ctx.conn.RemoteAddr() }

func (ctx *RequestCtx) LocalAddr() net.Addr { return ctx.conn.LocalAddr() }

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
	ctx.UserData.Reset()
	ctx.sessionOK = false
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

func (ctx *RequestCtx) setConnection(conn net.Conn) {
	ctx.r.Reset(conn)
	ctx.w.Reset(conn)
	ctx.connTime = time.Now()
	ctx.conn = conn
}

func (ctx *RequestCtx) Write(p []byte) (int, error) {
	return ctx.Response.Write(p)
}

func (ctx *RequestCtx) WriteString(s string) error {
	_, e := ctx.Write(utils.B(s))
	return e
}

func (ctx *RequestCtx) WriteJSON(v interface{}) error {
	ctx.Response.Header().SetContentType(MIMEJson)
	encoder := jsonx.NewEncoder(ctx)
	return encoder.Encode(v)
}

func (ctx *RequestCtx) WriteHTML(v []byte) error {
	ctx.Response.Header().SetContentType(MIMEHtml)
	_, e := ctx.Write(v)
	return e
}

func (ctx *RequestCtx) WriteStream(stream io.Reader) error {
	return sendChunkedStreamResponse(ctx.w, ctx, stream)
}

func (ctx *RequestCtx) WriteFile(f io.Reader, ext string) error {
	ctx.Response.Header().SetContentType(mime.TypeByExtension(ext))
	return ctx.WriteStream(f)
}

type Template interface {
	Execute(ctx context.Context, v interface{}) error
}

func (ctx *RequestCtx) WriteTemplate(t Template, data interface{}) error {
	ctx.Response.Header().SetContentType(MIMEHtml)
	return t.Execute(ctx, data)
}
