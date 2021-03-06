package sha

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"time"

	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
)

//RequestCtx most of the fields and most of the method return values are read-only, so:
// 1, if you want to modify them, you should keep this in mind.
// 2, if you want to keep them after handling, you should copy them.
type RequestCtx struct {
	noCopy
	ctx        context.Context
	cancelFunc func()
	readBuf    []byte
	conn       net.Conn
	r          *bufio.Reader
	w          *bufio.Writer
	connTime   time.Time

	Request  Request
	Response Response

	UserData userData
	err      interface{}
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

type WrappedError struct {
	A interface{}
	B interface{}
}

func (we *WrappedError) Error() string { return fmt.Sprintf("wrapped error: %v, %v", we.A, we.B) }

func (ctx *RequestCtx) SetError(v interface{}) {
	if v == nil {
		panic(errors.New("sha: nil"))
	}

	if ctx.err == nil {
		ctx.err = v
	} else {
		ctx.err = WrappedError{ctx.err, v}
	}
}

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

func (ctx *RequestCtx) IsTLS() bool { return ctx.Request.flags.Has(_ReqFlagIsTLS) }

func (ctx *RequestCtx) Conn() net.Conn { return ctx.conn }

func (ctx *RequestCtx) RemoteAddr() net.Addr { return ctx.conn.RemoteAddr() }

func (ctx *RequestCtx) LocalAddr() net.Addr { return ctx.conn.LocalAddr() }

func (ctx *RequestCtx) Hijack() net.Conn {
	ctx.Request.hijack()
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

func (ctx *RequestCtx) prepareForNextRequest(maxCap int) {
	ctx.Request.Reset(maxCap)
	ctx.Response.reset(maxCap)
	ctx.UserData.Reset()
	ctx.err = nil
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

func (ctx *RequestCtx) Reset(maxCap int) {
	if ctx.ctx == nil {
		return
	}
	ctx.prepareForNextRequest(maxCap)
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

type Message map[string]interface{}

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
