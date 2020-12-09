package suna

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/zzztttkkk/suna/internal"
	"golang.org/x/crypto/acme/autocert"
	"io"
	"log"
	"net"
	"time"
)

type Server struct {
	Host                   string
	Port                   int
	TlsConfig              *tls.Config
	BaseCtx                context.Context
	Handler                RequestHandler
	MaxConnectionKeepAlive time.Duration
	Http1xProtocol         Http1xProtocol
	AutoCompress           bool
	isTls                  bool
	beforeListen           []func(server *Server)
}

type _CtxVKey int

const (
	CtxServerKey = _CtxVKey(iota)
	CtxConnKey
)

func Default(handler RequestHandler) *Server {
	server := &Server{
		Host:                   "127.0.0.1",
		Port:                   8080,
		BaseCtx:                context.Background(),
		Handler:                handler,
		MaxConnectionKeepAlive: time.Minute * 3,
	}

	server.Http1xProtocol.Version = []byte("HTTP/1.1")
	server.Http1xProtocol.MaxFirstLintSize = 1024 * 4
	server.Http1xProtocol.MaxHeadersSize = 1024 * 8
	server.Http1xProtocol.MaxBodySize = 1024 * 1024 * 10
	server.Http1xProtocol.WriteTimeout = time.Second * 30
	server.Http1xProtocol.IdleTimeout = time.Second * 30
	server.Http1xProtocol.ReadBufferSize = 512

	server.beforeListen = append(
		server.beforeListen,
		func(server *Server) {
			go func() {
				time.Sleep(time.Second)

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
			}()
		},
	)

	return server
}

func (s *Server) doListen() net.Listener {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	log.Printf("suna: listening at `%s`\n", addr)

	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		panic(err)
	}
	return listener
}

//goland:noinspection GoSnakeCaseUsage  disable ide suggestion
func _0011111_strSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func (s *Server) enableTls(l net.Listener, certFile, keyFile string) net.Listener {
	if s.TlsConfig == nil {
		s.TlsConfig = &tls.Config{}
	}

	if !_0011111_strSliceContains(s.TlsConfig.NextProtos, "http/1.1") {
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

func (s *Server) doAccept(l net.Listener) {
	for _, fn := range s.beforeListen {
		fn(s)
	}

	s.Http1xProtocol.server = s
	s.Http1xProtocol.handler = s.Handler

	var tempDelay time.Duration
	ctx := context.WithValue(s.BaseCtx, CtxServerKey, s)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("suna.server: bad connection: %s\n", err.Error())
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
			}
			continue
		}
		if s.MaxConnectionKeepAlive > 0 {
			_ = conn.SetDeadline(time.Now().Add(s.MaxConnectionKeepAlive))
		}
		go s.serve(context.WithValue(ctx, CtxConnKey, conn), conn)
	}
}

func (s *Server) ListenAndServe() {
	s.doAccept(s.doListen())
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string) {
	s.isTls = true
	s.doAccept(s.enableTls(s.doListen(), certFile, keyFile))
}

func (s *Server) ListenAndServerWithAutoCert(hostnames ...string) {
	s.isTls = true
	s.doAccept(autocert.NewListener(hostnames...))
}

func (s *Server) tslHandshake(conn net.Conn) (*tls.Conn, string, error) {
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return nil, "", nil
	}
	err := tlsConn.Handshake()
	if err != nil {
		return nil, "", err
	}
	return tlsConn, tlsConn.ConnectionState().NegotiatedProtocol, nil
}

func (s *Server) serve(connCtx context.Context, conn net.Conn) {
	defer conn.Close()

	var subProtocolName string
	var tlsConn *tls.Conn
	var err error

	tlsConn, subProtocolName, err = s.tslHandshake(conn)
	if err != nil { // tls handshake error
		if re, ok := err.(tls.RecordHeaderError); ok && re.Conn != nil {
			_, _ = io.WriteString(
				re.Conn,
				"HTTP/1.0 400 Bad Request\r\n\r\nClient sent an HTTP request to an HTTPS server.\n",
			)
		}
		return
	}

	if tlsConn != nil {
		conn = tlsConn
		switch subProtocolName {
		case "", "http/1.0", "http/1.1":
		default:
			conn.Close()
		}
	}
	s.Http1xProtocol.Serve(connCtx, conn)
}
