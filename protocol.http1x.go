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
			//goland:noinspection GoNilness
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

				subProtocol, upgradeOk := rctx.Upgrade()
				if subProtocol == nil { // upgrade failed
					if rctx.Response.statusCode == 0 {
						protocol.server.Handler.Handle(rctx)
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
			}
		}
		continue
	}
}
