package sha

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/zzztttkkk/sha/internal"
	"golang.org/x/crypto/acme/autocert"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Server struct {
	Http1xProtocol

	Host      string
	Port      int
	TlsConfig *tls.Config
	BaseCtx   context.Context
	Handler   RequestHandler

	// connection
	MaxConnectionKeepAlive time.Duration
	ReadTimeout            time.Duration
	IdleTimeout            time.Duration
	WriteTimeout           time.Duration
	OnConnectionAccepted   func(conn net.Conn) bool

	AutoCompression bool
	isTls           bool
	beforeListen    []func(s *Server)
}

type _CtxVKey int

const (
	CtxKeyServer = _CtxVKey(iota)
	CtxKeyConnection
)

var prepareFunc []func(s *Server)

func init() {
	prepareFunc = append(prepareFunc,
		// setup recover
		func(server *Server) {
			_, ok := server.Handler.(*Mux)
			if !ok {
				return
			}
			for k, v := range internal.ErrorStatusByValue {
				RecoverByErr(
					k,
					func(sc int) ErrorHandler {
						return func(ctx *RequestCtx, _ interface{}) { ctx.SetStatus(sc) }
					}(v),
				)
			}
		},
		// init http1x_protocol
		func(server *Server) {
			protocol := &server.Http1xProtocol
			protocol.readBufferPool = internal.NewFixedSizeBufferPoll(
				protocol.ReadBufferSize, protocol.MaxReadBufferSize,
			)
			protocol.resBodyBufferPool = internal.NewBufferPoll(protocol.MaxResponseBodyBufferSize)
			protocol.resSendBufferPool = &sync.Pool{New: func() interface{} { return nil }}
		},
	)
}

func Default(handler RequestHandler) *Server {
	server := &Server{
		Host:                   "127.0.0.1",
		Port:                   5986,
		BaseCtx:                context.Background(),
		Handler:                handler,
		MaxConnectionKeepAlive: time.Minute * 3,
		WriteTimeout:           time.Second * 5,
		ReadTimeout:            time.Second * 10,
		IdleTimeout:            time.Second * 30,
	}

	server.Version = []byte("HTTP/1.1")
	server.MaxRequestFirstLineSize = 1024 * 4
	server.MaxRequestHeaderPartSize = 1024 * 8
	server.MaxRequestBodySize = 1024 * 1024 * 10
	server.ReadBufferSize = 512
	server.MaxReadBufferSize = 1024 * 4
	server.MaxResponseBodyBufferSize = 4096
	server.DefaultResponseSendBufferSize = 4096
	return server
}

func (s *Server) BeforeListening(fn func(s *Server)) {
	s.beforeListen = append(s.beforeListen, fn)
}

func (s *Server) doListen() net.Listener {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	log.Printf("sha: listening at `%s`\n", addr)

	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		panic(err)
	}
	return listener
}

func (s *Server) enableTls(l net.Listener, certFile, keyFile string) net.Listener {
	if s.TlsConfig == nil {
		s.TlsConfig = &tls.Config{}
	}

	if !internal.StrSliceContains(s.TlsConfig.NextProtos, "http/1.1") {
		s.TlsConfig.NextProtos = append(s.TlsConfig.NextProtos, "http/1.1")
	}
	configHasCert := len(s.TlsConfig.Certificates) > 0 || s.TlsConfig.GetCertificate != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		var err error
		s.TlsConfig.Certificates = make([]tls.Certificate, 1)
		s.TlsConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			panic(err)
		}
	}
	return tls.NewListener(l, s.TlsConfig)
}

func (s *Server) Serve(l net.Listener) {
	for _, fn := range prepareFunc {
		fn(s)
	}
	for _, fn := range s.beforeListen {
		fn(s)
	}
	s.BaseCtx = context.WithValue(s.BaseCtx, CtxKeyServer, s)

	s.Http1xProtocol.server = s
	s.Http1xProtocol.handler = s.Handler

	var tempDelay time.Duration
	var serveFunc func(conn net.Conn)
	serveFunc = s.serve
	if s.isTls {
		serveFunc = s.serveTLS
	}

	going := true
	go func() {
		for {
			select {
			case <-s.BaseCtx.Done():
				going = false
				_ = l.Close()
			}
		}
	}()

	for going {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("sha.server: bad connection: %s\n", err.Error())
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
			}
			continue
		}
		if s.OnConnectionAccepted != nil && !s.OnConnectionAccepted(conn) {
			conn.Close()
			continue
		}

		if s.MaxConnectionKeepAlive > 0 {
			_ = conn.SetDeadline(time.Now().Add(s.MaxConnectionKeepAlive))
		}
		go serveFunc(conn)
	}
}

func (s *Server) ListenAndServe() {
	s.Serve(s.doListen())
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string) {
	s.isTls = true
	s.Serve(s.enableTls(s.doListen(), certFile, keyFile))
}

func (s *Server) ListenAndServerWithAutoCert(hostnames ...string) {
	s.isTls = true
	s.Serve(autocert.NewListener(hostnames...))
}

var NonTLSRequestResponseMessage = `HTTP/1.0 400 Bad Request
Connection: close
Content-Length: 47

Client sent an HTTP request to an HTTPS server.`

var UnSupportedTLSSubProtocolRequestResponseMessage = `HTTP/1.0 510 Not Implemented
Connection: close
Content-Length: 25

UnSupportedTLSSubProtocol`

func (s *Server) serveTLS(conn net.Conn) {
	defer conn.Close()

	var tlsConn = conn.(*tls.Conn)
	var err error

	if s.ReadTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
	}
	err = tlsConn.Handshake()
	if err != nil { // tls handshake error
		if re, ok := err.(tls.RecordHeaderError); ok && re.Conn != nil {
			_, _ = io.WriteString(re.Conn, NonTLSRequestResponseMessage)
		}
		return
	}

	switch tlsConn.ConnectionState().NegotiatedProtocol {
	case "", "http/1.0", "http/1.1":
	default:
		_, _ = io.WriteString(tlsConn, UnSupportedTLSSubProtocolRequestResponseMessage)
		return
	}

	if s.ReadTimeout > 0 {
		_ = conn.SetReadDeadline(zeroTime)
	}
	s.Http1xProtocol.Serve(context.WithValue(s.BaseCtx, CtxKeyConnection, tlsConn), tlsConn)
}

func (s *Server) serve(conn net.Conn) {
	defer conn.Close()
	s.Http1xProtocol.Serve(context.WithValue(s.BaseCtx, CtxKeyConnection, conn), conn)
}
