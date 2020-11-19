package suna

import (
	"context"
	"net"
	"time"
)

type Http1xProtocol struct {
	Version          []byte
	MaxFirstLintSize int
	MaxHeadersSize   int
	MaxBodySize      int
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	OnParseError     func(conn net.Conn, err HttpError) bool
	OnWriteError     func(conn net.Conn, err error) bool
	ReadBufferSize   int
	SubProtocols     map[string]Protocol

	server *Server
	inited bool
}

func (protocol *Http1xProtocol) Handshake(ctx *RequestCtx) bool {
	return false
}

var closeStr = []byte("close")
var keepAliveStr = []byte("keep-alive")

func (protocol *Http1xProtocol) keepalive(ctx *RequestCtx) bool {
	if string(ctx.Request.version) < "1.1" {
		return false
	}
	connVal, _ := ctx.Request.Header.Get(headerConnection)
	if string(connVal) == "close" {
		return false
	}
	connVal, _ = ctx.Response.Header.Get(headerConnection)
	return string(connVal) != "close"
}

func (protocol *Http1xProtocol) Serve(ctx context.Context, conn net.Conn, _ *Request) {
	var err error
	var httpError HttpError
	var n int
	var stop bool

	rctx := AcquireRequestCtx()
	defer func() {
		ReleaseRequestCtx(rctx)
		_ = conn.Close()
	}()

	rctx.conn = conn
	rctx.connTime = time.Now()
	rctx.makeRequestCtx = func() context.Context { return ctx }
	rctx.protocol = protocol

	buf := make([]byte, protocol.ReadBufferSize)

	for !stop {
		select {
		case <-ctx.Done():
			{
				return
			}
			//goland:noinspection GoNilness
		default:
			offset := 0
			n, err = conn.Read(buf)
			if err != nil {
				return
			}

			for offset != n {
				offset, httpError = rctx.feedHttp1xReqData(buf, offset, n)
				if err != nil {
					stop = protocol.OnParseError(conn, httpError)
					break
				}
			}

			if rctx.status == 2 && rctx.bodyRemain < 1 {
				if rctx.Context == nil {
					rctx.initRequest()
				}

				keepalive := true
				subProtocol, upgradeOk := rctx.Upgrade()
				if subProtocol == nil { // upgrade failed
					if rctx.Response.statusCode == 0 {
						protocol.server.Handler.Handle(rctx)
					}
					keepalive = protocol.keepalive(rctx)
					if keepalive {
						rctx.Response.Header.Set(headerConnection, keepAliveStr)
					} else {
						rctx.Response.Header.Set(headerConnection, closeStr)
					}
				}

				if err := rctx.sendHttp1xResponseBuffer(); err != nil {
					stop = protocol.OnWriteError(conn, err)
				}

				if upgradeOk {
					//goland:noinspection GoNilness
					subProtocol.Serve(ctx, conn, &rctx.Request)
					return
				}
				rctx.Reset()

				if !keepalive {
					return
				}
			}
		}
		continue
	}
}
