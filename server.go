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
	beforeListen    []func()
}

type _CtxVKey int

const (
	CtxKeyServer = _CtxVKey(iota)
	CtxKeyConnection
)

func Default(handler RequestHandler) *Server {
	server := &Server{
		Host:                   "127.0.0.1",
		Port:                   8080,
		BaseCtx:                context.Background(),
		Handler:                handler,
		MaxConnectionKeepAlive: time.Minute * 3,
		WriteTimeout:           time.Second * 5,
		ReadTimeout:            time.Second * 10,
		IdleTimeout:            time.Second * 30,
	}

	server.Http1xProtocol.Version = []byte("HTTP/1.1")
	server.Http1xProtocol.MaxRequestFirstLineSize = 1024 * 4
	server.Http1xProtocol.MaxRequestHeaderPartSize = 1024 * 8
	server.Http1xProtocol.MaxRequestBodySize = 1024 * 1024 * 10
	server.Http1xProtocol.ReadBufferSize = 512
	server.Http1xProtocol.MaxResponseBodyBufferSize = 4096
	server.Http1xProtocol.DefaultResponseSendBufferSize = 4096

	server.beforeListen = append(
		server.beforeListen,
		func() {
			mux, ok := server.Handler.(*_Mux)
			if !ok {
				return
			}
			for k, v := range internal.ErrorStatusByValue {
				mux.RecoverByErr(
					k,
					func(sc int) ErrorHandler {
						return func(ctx *RequestCtx, _ interface{}) { ctx.SetStatus(sc) }
					}(v),
				)
			}
		},
		func() {
			protocol := &server.Http1xProtocol
			protocol.readBufferPool = internal.NewFixedSizeBufferPoll(
				protocol.ReadBufferSize, protocol.MaxReadBufferSize,
			)
			protocol.resBodyBufferPool = internal.NewBufferPoll(protocol.MaxResponseBodyBufferSize)
			protocol.resSendBufferPool = &sync.Pool{New: func() interface{} { return nil }}
		},
	)

	return server
}

func (s *Server) BeforeListening(fn func()) {
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
	for _, fn := range s.beforeListen {
		fn()
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

var NonTLSRequestResponseMessage = "HTTP/1.0 400 Bad Request\n\nClient sent an HTTP request to an HTTPS server.\n"

func (s *Server) serveTLS(conn net.Conn) {
	defer conn.Close()

	var tlsConn *tls.Conn
	var err error
	var ok bool

	if s.ReadTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
	}

	tlsConn, ok = conn.(*tls.Conn)
	if !ok {
		_, _ = io.WriteString(conn, NonTLSRequestResponseMessage)
		return
	}

	err = tlsConn.Handshake()
	if err != nil { // tls handshake error
		if re, ok := err.(tls.RecordHeaderError); ok && re.Conn != nil {
			_, _ = io.WriteString(re.Conn, NonTLSRequestResponseMessage)
		}
		return
	}

	if tlsConn != nil {
		conn = tlsConn
		switch tlsConn.ConnectionState().NegotiatedProtocol {
		case "", "http/1.0", "http/1.1":
		default:
			conn.Close()
		}
	}

	if s.ReadTimeout > 0 {
		_ = conn.SetReadDeadline(zeroTime)
	}
	s.Http1xProtocol.Serve(context.WithValue(s.BaseCtx, CtxKeyConnection, conn), conn)
}

func (s *Server) serve(conn net.Conn) {
	defer conn.Close()
	s.Http1xProtocol.Serve(context.WithValue(s.BaseCtx, CtxKeyConnection, conn), conn)
}
