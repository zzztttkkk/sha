package sha

import (
	"context"
	"crypto/tls"
	"github.com/imdario/mergo"
	"github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/utils"
	"github.com/zzztttkkk/websocket"
	"golang.org/x/crypto/acme/autocert"
	"io"
	"log"
	"net"
	"time"
)

type ServerOption struct {
	Addr string `json:"addr" toml:"addr"`
	TLS  struct {
		AutoCertDomains []string `json:"auto_cert_domains" toml:"auto-cert-domains"`
		Key             string   `json:"key" toml:"key"`
		Cert            string   `json:"cert" toml:"cert"`
	} `json:"tls" toml:"tls"`
	MaxConnectionKeepAlive utils.TomlDuration `json:"max_connection_keep_alive" toml:"max-connection-keep-alive"`
	ReadTimeout            utils.TomlDuration `json:"read_timeout" toml:"read-timeout"`
	IdleTimeout            utils.TomlDuration `json:"idle_timeout" toml:"idle-timeout"`
	WriteTimeout           utils.TomlDuration `json:"write_timeout" toml:"write-timeout"`

	HTTPProtocol      HTTPProtocol      `json:"-" toml:"-"`
	WebsocketProtocol WebSocketProtocol `json:"-" toml:"-"`
}

var defaultServerOption = ServerOption{
	Addr:                   "127.0.0.1:5986",
	MaxConnectionKeepAlive: utils.TomlDuration{Duration: time.Minute * 5},
}

type Server struct {
	option      ServerOption
	readTimeout time.Duration

	OnConnectionAccepted func(conn net.Conn) bool

	baseCtx           context.Context
	Handler           RequestHandler
	httpProtocol      HTTPProtocol
	websocketProtocol WebSocketProtocol

	tls          *tls.Config
	isTLS        bool
	beforeAccept []func(s *Server)
}

func (s *Server) IsTLS() bool { return s.isTLS }

type HTTPProtocol interface {
	ServeHTTPConn(ctx context.Context, conn net.Conn)
}

type WebSocketProtocol interface {
	Handshake(ctx *RequestCtx) bool
	Hijack(ctx *RequestCtx) *websocket.Conn
}

type _CtxVKey int

const (
	CtxKeyServer = _CtxVKey(iota)
	CtxKeyConnection
)

var serverPrepareFunc []func(s *Server)

func init() {
	serverPrepareFunc = append(
		serverPrepareFunc,
		func(server *Server) {
			hp, ok := server.httpProtocol.(*_Http11Protocol)
			if ok {
				hp.server = server
			}
		},
	)
}

func New(ctx context.Context, opt *ServerOption) *Server {
	if opt == nil {
		_v := &ServerOption{}
		*_v = defaultServerOption
		opt = _v
	}

	if opt.HTTPProtocol == nil {
		opt.HTTPProtocol = NewHTTP11Protocol(nil)
	}

	if opt.WebsocketProtocol == nil {
		opt.WebsocketProtocol = NewWebSocketProtocol(nil)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	server := &Server{httpProtocol: opt.HTTPProtocol, websocketProtocol: opt.WebsocketProtocol, baseCtx: ctx}
	server.option = *opt

	if err := mergo.Merge(&server.option, &defaultServerOption); err != nil {
		panic(err)
	}

	server.readTimeout = server.option.ReadTimeout.Duration
	return server
}

func (s *Server) BeforeAccept(fn func(s *Server)) {
	s.beforeAccept = append(s.beforeAccept, fn)
}

func (s *Server) doListen() net.Listener {
	log.Printf("sha: listening at `%s`\n", s.option.Addr)

	listener, err := net.Listen("tcp4", s.option.Addr)
	if err != nil {
		panic(err)
	}
	return listener
}

func (s *Server) enableTLS(l net.Listener, certFile, keyFile string) net.Listener {
	if certFile == "" || keyFile == "" {
		panic("sha: empty tls file")
	}

	if s.tls == nil {
		s.tls = &tls.Config{}
	}

	if !internal.StrSliceContains(s.tls.NextProtos, "http/1.1") {
		s.tls.NextProtos = append(s.tls.NextProtos, "http/1.1")
	}
	configHasCert := len(s.tls.Certificates) > 0 || s.tls.GetCertificate != nil
	if !configHasCert {
		var err error
		s.tls.Certificates = make([]tls.Certificate, 1)
		s.tls.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			panic(err)
		}
	}
	return tls.NewListener(l, s.tls)
}

func (s *Server) serve(l net.Listener) {
	for _, fn := range serverPrepareFunc {
		fn(s)
	}
	for _, fn := range s.beforeAccept {
		fn(s)
	}
	s.baseCtx = context.WithValue(s.baseCtx, CtxKeyServer, s)

	var tempDelay time.Duration
	var serveFunc func(conn net.Conn)
	serveFunc = s.serveConn
	if s.isTLS {
		serveFunc = s.serveTLS
	}

	f := true
	go func() {
		for {
			select {
			case <-s.baseCtx.Done():
				f = false
				_ = l.Close()
			}
		}
	}()

	maxKeepAlive := s.option.MaxConnectionKeepAlive.Duration

	for f {
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
			_ = conn.Close()
			continue
		}

		if maxKeepAlive > 0 {
			_ = conn.SetDeadline(time.Now().Add(maxKeepAlive))
		}
		go serveFunc(conn)
	}
}

func (s *Server) ListenAndServe() {
	if len(s.option.TLS.AutoCertDomains) > 0 {
		s.isTLS = true
		s.serve(autocert.NewListener(s.option.TLS.AutoCertDomains...))
		return
	}

	if len(s.option.TLS.Cert) > 0 {
		s.isTLS = true
		s.serve(s.enableTLS(s.doListen(), s.option.TLS.Cert, s.option.TLS.Key))
		return
	}

	s.serve(s.doListen())
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

	if s.readTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(s.readTimeout))
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

	if s.readTimeout > 0 {
		_ = conn.SetReadDeadline(zeroTime)
	}
	s.httpProtocol.ServeHTTPConn(context.WithValue(s.baseCtx, CtxKeyConnection, tlsConn), tlsConn)
}

func (s *Server) serveConn(conn net.Conn) {
	defer conn.Close()
	s.httpProtocol.ServeHTTPConn(context.WithValue(s.baseCtx, CtxKeyConnection, conn), conn)
}

func ListenAndServe(addr string, handler RequestHandler) {
	if addr == "" {
		addr = "127.0.0.1:5986"
	}

	server := New(context.Background(), nil)
	server.option.Addr = addr
	server.Handler = handler
	server.ListenAndServe()
}
