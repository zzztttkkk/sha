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

	server  *Server
	handler RequestHandler
}

func (protocol *Http1xProtocol) Handshake(_ *RequestCtx) bool {
	return false
}

var closeStr = []byte("close")
var upgradeStr = []byte("upgrade")
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
var MaxIdleSleepTime = time.Millisecond * 100

func (protocol *Http1xProtocol) Serve(ctx context.Context, _ *Request, conn net.Conn) {
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

	sleepDu := time.Millisecond * 10
	resetIdleTimeout := true

	for !stop {
		select {
		case <-ctx.Done():
			{
				return
			}
			//goland:noinspection GoNilness
		default:
			if protocol.IdleTimeout > 0 && resetIdleTimeout {
				conn.SetReadDeadline(time.Now().Add(protocol.IdleTimeout))
			}

			offset := 0
			n, err = conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					time.Sleep(sleepDu)
					sleepDu = time.Duration(int64(float64(sleepDu) * 1.5))
					resetIdleTimeout = false
					if sleepDu > MaxIdleSleepTime {
						sleepDu = MaxIdleSleepTime
					}
					continue
				}
				return
			}

			if protocol.IdleTimeout > 0 {
				conn.SetReadDeadline(ZeroTime)
				resetIdleTimeout = true
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

				if protocol.WriteTimeout > 0 {
					_ = conn.SetWriteDeadline(time.Now().Add(protocol.WriteTimeout))
				}

				subp := rctx.UpgradeProtocol()
				if subp != nil {
					if subp.Handshake(rctx) {
						subp.Serve(rctx.Context, &rctx.Request, conn)
						return
					}
				}

				protocol.handler.Handle(rctx)

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

				rctx.Reset()
			}
		}
		continue
	}
}
