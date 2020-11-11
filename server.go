package suna

import (
	"context"
	"fmt"
	"net"
	"time"
)

type HttpProtocol struct {
	MaxFirstLintSize int
	MaxHeadersSize   int
	MaxBodySize      int
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
}

type Server struct {
	HttpProtocol      HttpProtocol
	Host              string
	Port              int
	BaseCtx           context.Context
	Handler           RequestHandler
	OnConnectionError func(conn net.Conn, err error)
}

type _Key int

const (
	CtxServerKey = iota
	CtxConnKey
)

func (s *Server) Run() {
	listener, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		panic(err)
	}

	var tempDelay time.Duration
	ctx := context.WithValue(s.BaseCtx, CtxServerKey, s)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
		}

		go s.serve(context.WithValue(ctx, CtxConnKey, conn), conn)
	}
}

var ConnReadBufferSize = 128 // min mem required for each conn

func (s *Server) onParseErr(conn net.Conn, e error) {
	if s.OnConnectionError != nil {
		s.OnConnectionError(conn, e)
		return
	}

	var msg []byte
	switch e {
	case ErrRequestHeaderFieldsTooLarge:
		msg = []byte("HTTP/1.0 431 Request Header Fields Too Large\r\n\r\n")
	case ErrPayloadTooLarge:
		msg = []byte("HTTP/1.0 413 Payload Too Large\r\n\r\n")
	case ErrBadConnection:
		msg = []byte("HTTP/1.0 503 Service Unavailable\r\n\r\n")
	case ErrUrlTooLong:
		msg = []byte("HTTP/1.0 414 URI Too Long\r\n\r\n")
	}
	_, _ = conn.Write(msg)
}

func (s *Server) serve(ctx context.Context, conn net.Conn) {
	rctx := AcquireRequestCtx()
	defer func() {
		ReleaseRequestCtx(rctx)
		conn.Close()
	}()

	buf := make([]byte, ConnReadBufferSize)

	rctx.conn = conn
	rctx.connTime = time.Now()
	var err error
	var n int

	rctx.ctxFun = func() context.Context { return ctx }

	for {
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
				offset, err = rctx.feedReqData(buf, offset, n, &s.HttpProtocol)
				if err != nil {
					s.onParseErr(conn, err)
					return
				}
			}

			if rctx.status == 2 && rctx.bodyRemain < 1 {
				if rctx.Context == nil {
					rctx.initRequest()
				}
				s.Handler.Handle(rctx)
				rctx.prepareForNext()
			}
			continue
		}
	}
}
