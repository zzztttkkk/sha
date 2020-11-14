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

	inited bool
}

func (protocol *Http1xProtocol) Upgrade(
	connCtx context.Context,
	conn net.Conn, ctx *RequestCtx, name []byte,
) Protocol {
	return nil
}

func (protocol *Http1xProtocol) Serve(ctx context.Context, s *Server, conn net.Conn) {
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

				upgrade := rctx.UpgradeTo()
				if len(upgrade) > 0 {
					// handshake
					nProtocol := s.protocol.Upgrade(ctx, conn, rctx, upgrade)
					if nProtocol == nil {
						return
					} else {
						nProtocol.Serve(ctx, s, conn)
						return
					}
				}

				s.Handler.Handle(rctx)
				if err := rctx.sendHttp1xResponseBuffer(); err != nil {
					stop = protocol.OnWriteError(conn, err)
				}
				rctx.reset()
			}
			continue
		}
	}
}
