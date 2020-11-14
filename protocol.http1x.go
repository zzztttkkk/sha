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

func (protocol *Http1xProtocol) Serve(ctx context.Context, conn net.Conn, _ *Request) {
	var err error
	var herr HttpError
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
		default:
			offset := 0
			n, err = conn.Read(buf)
			if err != nil {
				return
			}

			for offset != n {
				offset, herr = rctx.feedHttp1xReqData(buf, offset, n)
				if err != nil {
					stop = protocol.OnParseError(conn, herr)
					break
				}
			}

			if rctx.status == 2 && rctx.bodyRemain < 1 {
				if rctx.Context == nil {
					rctx.initRequest()
				}

				subProtocol, ok := rctx.Upgrade()
				if subProtocol == nil { // not an upgrade request
					protocol.server.Handler.Handle(rctx)
				} else {
					_ = rctx.sendHttp1xResponseBuffer() // send upgrade response
					if ok {
						subProtocol.Serve(ctx, conn, &rctx.Request)
						return
					}
				}
				if err := rctx.sendHttp1xResponseBuffer(); err != nil {
					stop = protocol.OnWriteError(conn, err)
				}
				rctx.reset()
			}
		}
		continue
	}
}
