package sha

import (
	"context"
	"github.com/zzztttkkk/sha/utils"
	"net"
	"time"
)

type HTTPOptions struct {
	MaxFirstLineSize  int `json:"max_first_line_size" toml:"max-first-line-size"`
	MaxHeaderPartSize int `json:"max_header_part_size" toml:"max-header-part-size"`
	MaxBodySize       int `json:"max_body_size" toml:"max-body-size"`

	ReadBufferSize      int `json:"read_buffer_size" toml:"read-buffer-size"`
	SendBufferSize      int `json:"send_buffer_size" toml:"send-buffer-size"`
	BufferPoolSizeLimit int `json:"buffer_pool_size_limit" toml:"buffer-pool-size-limit"`
}

var defaultHTTPOption = HTTPOptions{
	MaxFirstLineSize:    4096,
	MaxHeaderPartSize:   1024 * 16,
	MaxBodySize:         1024 * 1024 * 10,
	ReadBufferSize:      2048,
	BufferPoolSizeLimit: 2048,
}

type _Http11Protocol struct {
	HTTPOptions

	OnParseError func(conn net.Conn, err error) bool                  // keep connection if return true
	OnWriteError func(conn net.Conn, ctx *RequestCtx, err error) bool // keep connection if return true

	ReadBufferSize  int
	WriteBufferSize int

	pool *RequestCtxPool
}

func newHTTP11Protocol(pool *RequestCtxPool) HTTPServerProtocol {
	option := pool.opt
	v := &_Http11Protocol{HTTPOptions: *option}
	if v.BufferPoolSizeLimit > v.MaxBodySize {
		v.BufferPoolSizeLimit = v.MaxBodySize
	}

	if pool == nil {
		pool = defaultRCtxPool
	}
	v.pool = pool
	return v
}

const (
	headerValClose = "close"
	keepAlive      = "keep-alive"
	upgrade        = "upgrade"
	five           = time.Second * 5
)

func (protocol *_Http11Protocol) keepalive(ctx *RequestCtx, s *Server) bool {
	if !s.runningFlag {
		return false
	}
	timeout := s.Options.MaxConnectionKeepAlive.Duration
	if timeout > 0 && timeout-time.Now().Sub(ctx.connTime) <= five {
		return false
	}

	connVal, _ := ctx.Response.Header().Get(HeaderConnection) // disable keep-alive by response
	if len(connVal) > 0 && utils.S(connVal) == headerValClose {
		return false
	}

	connVal, _ = ctx.Request.Header().Get(HeaderConnection) // disable keep-alive by request
	if len(connVal) > 0 && utils.S(inPlaceLowercase(connVal)) == headerValClose {
		return false
	}

	v := ctx.Request.HTTPVersion()
	return v[5] >= '1' && v[7] >= '1'
}

func (protocol *_Http11Protocol) handle(ctx *RequestCtx, server *Server) bool {
	defer func() {
		ctx.cancelFunc()
		ctx.prepareForNextRequest()
	}()

	readTimeout := server.Options.ReadTimeout.Duration
	if readTimeout > 0 {
		_ = ctx.conn.SetReadDeadline(time.Now().Add(readTimeout))
	}

	err := parseRequest(ctx, ctx.r, ctx.readBuf, &ctx.Request, &protocol.HTTPOptions)
	if err != nil {
		if protocol.OnParseError != nil {
			return protocol.OnParseError(ctx.conn, err)
		}
		return false
	}

	server.Handler.Handle(ctx)
	if ctx.hijacked { // another protocol process has been completed
		return false
	}
	shouldKeepAlive := protocol.keepalive(ctx, server)

	writeTimeout := server.Options.WriteTimeout.Duration
	if writeTimeout > 0 {
		_ = ctx.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	}
	if err := sendResponse(ctx.w, &ctx.Response); err != nil {
		if protocol.OnWriteError != nil {
			return protocol.OnWriteError(ctx.conn, ctx, err)
		}
		return false
	}
	if writeTimeout > 0 {
		_ = ctx.conn.SetWriteDeadline(time.Time{})
	}
	return shouldKeepAlive
}

func (protocol *_Http11Protocol) ServeConn(ctx context.Context, conn net.Conn) {
	var shouldKeepAlive = true
	server := ctx.Value(CtxKeyServer).(*Server)

	rctx := protocol.pool.Acquire()
	rctx.isTLS = server.isTLS
	defer protocol.pool.release(rctx, false)

	rctx.setConnection(conn)
	idleTimeout := server.Options.IdleTimeout.Duration

	for shouldKeepAlive {
		rctx.ctx, rctx.cancelFunc = context.WithCancel(ctx)
		shouldKeepAlive = protocol.handle(rctx, server)
		if !shouldKeepAlive {
			return
		}

		if idleTimeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(idleTimeout))
			if _, err := rctx.r.Peek(4); err != nil {
				return
			}
		}
		_ = conn.SetReadDeadline(time.Time{})
		_ = conn.SetWriteDeadline(time.Time{})
	}
}
