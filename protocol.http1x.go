package suna

import (
	"context"
	"github.com/zzztttkkk/suna/internal"
	"io"
	"net"
	"time"
)

type Http1xProtocol struct {
	Version                  []byte
	MaxRequestFirstLineSize  int
	MaxRequestHeaderPartSize int
	MaxRequestBodySize       int

	IdleTimeout  time.Duration
	WriteTimeout time.Duration
	OnParseError func(conn net.Conn, err HttpError) (shouldCloseConn bool) // close connection if return true
	OnWriteError func(conn net.Conn, err error) (shouldCloseConn bool)     // close connection if return true

	ReadBufferSize     int
	MaxReadBufferSize  int
	MaxWriteBufferSize int

	server  *Server
	handler RequestHandler

	readBufferPool  *internal.FixedSizeBufferPool
	writeBufferPool *internal.BufferPool
}

var upgradeStr = []byte("upgrade")
var keepAliveStr = []byte("keep-alive")

const (
	closeStr  = "close"
	http11Str = "1.1"
)

func (protocol *Http1xProtocol) keepalive(ctx *RequestCtx) bool {
	req := &ctx.Request
	if string(req.version[5:]) < http11Str {
		return false
	}

	connVal, _ := ctx.Request.Header.Get(internal.B(HeaderConnection))
	if string(inplaceLowercase(connVal)) == closeStr {
		return false
	}
	connVal, _ = ctx.Response.Header.Get(internal.B(HeaderConnection))
	return string(inplaceLowercase(connVal)) != closeStr
}

var ZeroTime time.Time
var MaxIdleSleepTime = time.Millisecond * 100

func (protocol *Http1xProtocol) Serve(ctx context.Context, conn net.Conn) {
	var err error
	var n int
	var stop bool

	rctx := AcquireRequestCtx()
	readBuf := protocol.readBufferPool.Get()
	rctx.Response.buf = protocol.writeBufferPool.Get()

	defer func() {
		protocol.readBufferPool.Put(readBuf)
		protocol.writeBufferPool.Put(rctx.Response.buf)
		ReleaseRequestCtx(rctx)
	}()

	rctx.conn = conn
	rctx.connTime = time.Now()
	rctx.Context = ctx
	rctx.protocol = protocol

	sleepDu := time.Millisecond * 10
	resetIdleTimeout := true

	for !stop {
		select {
		case <-ctx.Done():
			{
				return
			}
		default:
			if protocol.IdleTimeout > 0 && resetIdleTimeout {
				conn.SetReadDeadline(time.Now().Add(protocol.IdleTimeout))
			}

			offset := 0
			n, err = conn.Read(readBuf.Data)
			if err != nil {
				if err == io.EOF {
					time.Sleep(sleepDu)
					sleepDu = sleepDu * 2
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
				offset, err = rctx.feedHttp1xReqData(readBuf.Data, offset, n)
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
				if protocol.server.AutoCompress {
					rctx.AutoCompress()
				}

				protocol.handler.Handle(rctx)

				if rctx.hijacked {
					return
				}

				if protocol.WriteTimeout > 0 {
					_ = conn.SetWriteDeadline(time.Now().Add(protocol.WriteTimeout))
				}

				if protocol.keepalive(rctx) {
					rctx.Response.Header.Set(internal.B(HeaderConnection), keepAliveStr)
				} else {
					stop = true
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

				rctx.Reset()
			}
		}
		continue
	}
}
