package sha

import (
	"context"
	"errors"
	"github.com/zzztttkkk/sha/utils"
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

	protocol *Http1xProtocol

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
}

type _Key int

const (
	_RCtxKey = _Key(iota + 3)
)

func (ctx *RequestCtx) Context() context.Context { return context.WithValue(ctx.ctx, _RCtxKey, ctx) }

func (ctx *RequestCtx) SetData(key string, value interface{}) { ctx.ud.Set(key, value) }

func (ctx *RequestCtx) GetData(key string) interface{} { return ctx.ud.Get(key) }

var ErrBadContext = errors.New("sha: bad context")

func RCtx(ctx context.Context) *RequestCtx {
	v := ctx.Value(_RCtxKey)
	if v == nil {
		panic(ErrBadContext)
	}
	return v.(*RequestCtx)
}

type MutexRequestCtx struct {
	sync.Mutex
	*RequestCtx
}

func (ctx *RequestCtx) IsTLS() bool { return ctx.isTLS }

func (ctx *RequestCtx) RemoteAddr() net.Addr {
	return ctx.conn.RemoteAddr()
}

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
	if string(inplaceLowercase(v)) != lowerUpgradeHeader {
		return ""
	}
	v, ok = ctx.Request.Header.Get(HeaderUpgrade)
	if !ok {
		return ""
	}
	return utils.S(inplaceLowercase(v))
}

func (ctx *RequestCtx) Reset() {
	if ctx.ctx == nil {
		return
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
