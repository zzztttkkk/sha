package suna

import (
	"context"
	"io"
	"net"
	"time"
)

type Http1xProtocol struct {
	Version          []byte
	MaxFirstLintSize int
	MaxHeadersSize   int
	MaxBodySize      int
	IdleTimeout      time.Duration
	WriteTimeout     time.Duration

	OnParseError   func(conn net.Conn, err HttpError) bool // close connection if return true
	OnWriteError   func(conn net.Conn, err error) bool     // close connection if return true
	ReadBufferSize int
	SubProtocols   map[string]Protocol

	server *Server
}

func (protocol *Http1xProtocol) Handshake(_ *RequestCtx) bool {
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

var ZeroTime time.Time

func (protocol *Http1xProtocol) Serve(ctx context.Context, conn net.Conn, _ *Request) {
	var err error
	var n int
	var stop bool

	rctx := AcquireRequestCtx()
	defer ReleaseRequestCtx(rctx)

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
			if protocol.IdleTimeout > 0 {
				conn.SetReadDeadline(time.Now().Add(protocol.IdleTimeout))
			}

			offset := 0
			n, err = conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					time.Sleep(time.Millisecond * 20)
					continue
				}
				return
			}

			if protocol.IdleTimeout > 0 {
				conn.SetReadDeadline(ZeroTime)
			}

			for offset != n {
				offset, err = rctx.feedHttp1xReqData(buf, offset, n)
				if err != nil {
					if protocol.OnParseError != nil {
						if protocol.OnParseError(conn, err.(HttpError)) {
							return
						}
					} else {
						return
					}
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

				if protocol.WriteTimeout > 0 {
					_ = conn.SetWriteDeadline(time.Now().Add(protocol.WriteTimeout))
				}

				if err := rctx.sendHttp1xResponseBuffer(); err != nil {
					if protocol.OnWriteError != nil {
						if protocol.OnWriteError(conn, err) {
							return
						}
					} else {
						return
					}
				}

				if protocol.WriteTimeout > 0 {
					_ = conn.SetWriteDeadline(ZeroTime)
				}

				if upgradeOk {
					//goland:noinspection GoNilness
					subProtocol.Serve(ctx, conn, &rctx.Request)
					return
				}

				rctx.Reset()
			}
		}
		continue
	}
}
