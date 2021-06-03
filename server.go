package sha

import (
	"context"
	"crypto/tls"
	"github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/utils"
	"github.com/zzztttkkk/websocket"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type ServerOptions struct {
	Network string `json:"network" toml:"network"`
	Addr    string `json:"addr" toml:"addr"`
	TLS     struct {
		AutoCertDomains []string `json:"auto_cert_domains" toml:"auto-cert-domains"`
		Key             string   `json:"key" toml:"key"`
		Cert            string   `json:"cert" toml:"cert"`
	} `json:"tls" toml:"tls"`

	MaxConnectionKeepAlive utils.TomlDuration `json:"max_connection_keep_alive" toml:"max-connection-keep-alive"`
	ReadTimeout            utils.TomlDuration `json:"read_timeout" toml:"read-timeout"`
	IdleTimeout            utils.TomlDuration `json:"idle_timeout" toml:"idle-timeout"`
	WriteTimeout           utils.TomlDuration `json:"write_timeout" toml:"write-timeout"`

	GracefullyShutdown bool   `json:"graceful_shutdown" toml:"graceful_shutdown"` //shut down until all connections are closed
	Pid                string `json:"pid" toml:"pid"`                             //pid file path
}

var defaultServerOption = ServerOptions{
	Network:                "tcp4",
	Addr:                   "127.0.0.1:5986",
	MaxConnectionKeepAlive: utils.TomlDuration{Duration: time.Minute * 5},
	ReadTimeout:            utils.TomlDuration{Duration: time.Second * 10},
	IdleTimeout:            utils.TomlDuration{Duration: time.Second * 30},
	WriteTimeout:           utils.TomlDuration{Duration: time.Second * 10},
}

type Server struct {
	Options ServerOptions

	OnConnectionAccepted func(conn net.Conn) bool

	baseCtx           context.Context
	Handler           RequestHandler
	httpProtocol      HTTPServerProtocol
	websocketProtocol WebSocketProtocol

	tls   *tls.Config
	isTLS bool

	// lifecycle
	beforeAccept   []func(s *Server)
	beforeShutdown []func(s *Server)
	aliveConns     int64
	runningFlag    bool
	listener       net.Listener
	shutdownOnce   sync.Once
	shutdownChan   *chan struct{}

	pool *RequestCtxPool
}

func (s *Server) IsTLS() bool { return s.isTLS }

type HTTPServerProtocol interface {
	ServeConn(ctx context.Context, conn net.Conn)
}

type WebSocketProtocol interface {
	Handshake(ctx *RequestCtx) (string, bool, bool)
	Hijack(ctx *RequestCtx, subProtocol string, compress bool) *websocket.Conn
}

type _CtxVKey int

const (
	CtxKeyServer = _CtxVKey(iota)
	CtxKeyConnection
	CtxKeyRequestCtx
)

var serverPrepareFunc []func(s *Server)

func New(ctx context.Context, pool *RequestCtxPool, opt *ServerOptions) *Server {
	if opt == nil {
		opt = &defaultServerOption
	}

	if pool == nil {
		pool = defaultRCtxPool
	}

	if ctx == nil {
		ctx = context.Background()
	}
	return &Server{
		Options: *opt,
		baseCtx: ctx,
		pool:    pool,
	}
}

func (s *Server) BeforeShutdown(fn func(server *Server)) {
	s.beforeShutdown = append(s.beforeShutdown, fn)
}

func (s *Server) SetHTTPProtocol(protocol HTTPServerProtocol) { s.httpProtocol = protocol }

func (s *Server) SetWebSocketProtocol(protocol WebSocketProtocol) { s.websocketProtocol = protocol }

func Default() *Server { return New(nil, nil, nil) }

func DefaultWithContext(ctx context.Context) *Server { return New(ctx, nil, nil) }

func (s *Server) RequestCtxPool() *RequestCtxPool { return s.pool }

func (s *Server) BeforeAccept(fn func(s *Server)) {
	s.beforeAccept = append(s.beforeAccept, fn)
}

func (s *Server) Listen() net.Listener {
	if s.Handler == nil {
		s.Handler = RequestHandlerFunc(func(ctx *RequestCtx) { _ = ctx.WriteString("Hello World!\n") })
	}

	listener, err := net.Listen(s.Options.Network, s.Options.Addr)
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

func (s *Server) Shutdown() {
	s.shutdownOnce.Do(func() {
		sc := make(chan struct{}, 1)
		s.shutdownChan = &sc

		s.runningFlag = false
		_ = s.listener.Close()

		if s.Options.GracefullyShutdown {
			if s.aliveConns > 0 {
				log.Printf(
					"sha.server: shutdown waiting, alive connections(%d); Pid: %d\r\n",
					s.aliveConns, os.Getpid(),
				)
			}
			for {
				if atomic.LoadInt64(&s.aliveConns) < 1 {
					break
				} else {
					time.Sleep(time.Millisecond * 100)
				}
			}
		}

		for _, fn := range s.beforeShutdown {
			fn(s)
		}
		if len(s.Options.Pid) > 0 {
			if e := os.Remove(s.Options.Pid); e != nil {
				log.Printf("sha.server: shutdown error, %s\r\n", e.Error())
			}
		}
		log.Printf("sha.server: shutdown done; Pid: %d\r\n", os.Getpid())
		sc <- struct{}{}
	})
}

func (s *Server) Serve(l net.Listener) {
	s.listener = l
	s.runningFlag = true

	if s.httpProtocol == nil {
		s.httpProtocol = newHTTP11Protocol(s.pool)
	}

	if s.websocketProtocol == nil {
		s.websocketProtocol = NewWebSocketProtocol(nil)
	}

	for _, fn := range serverPrepareFunc {
		fn(s)
	}
	for _, fn := range s.beforeAccept {
		fn(s)
	}
	s.baseCtx = context.WithValue(s.baseCtx, CtxKeyServer, s)

	var tempDelay time.Duration
	serve := s.serveHTTPConn
	if s.isTLS {
		serve = s.serveTLS
	}

	go func() {
		<-s.baseCtx.Done()
		s.Shutdown()
	}()

	maxKeepAlive := s.Options.MaxConnectionKeepAlive.Duration

	if len(s.Options.Pid) > 0 {
		pid, e := os.OpenFile(s.Options.Pid, os.O_CREATE|os.O_WRONLY, 0600)
		if e != nil {
			panic(e)
		}
		_, _ = pid.WriteString(strconv.FormatInt(int64(os.Getpid()), 10))
		_ = pid.Close()
	}

	log.Printf("sha.server: listening @ `%s`, Pid: %d\r\n", s.Options.Addr, os.Getpid())

	for s.runningFlag {
		conn, err := l.Accept()
		if err != nil {
			if !s.runningFlag {
				break
			}
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

		if maxKeepAlive > 0 {
			_ = conn.SetDeadline(time.Now().Add(maxKeepAlive))
		}

		go func(c net.Conn) {
			if s.OnConnectionAccepted != nil && !s.OnConnectionAccepted(c) {
				return
			}

			atomic.AddInt64(&s.aliveConns, 1)
			defer func() {
				_ = c.Close()
				atomic.AddInt64(&s.aliveConns, -1)
			}()

			serve(c)
		}(conn)
	}

	// waiting for shutdown
	if s.shutdownChan == nil {
		s.Shutdown()
	}
	<-(*s.shutdownChan)
	log.Printf("sha.server: stop @ `%s`, Pid: %d\r\n", s.Options.Addr, os.Getpid())
}

func (s *Server) ListenAndServe() {
	if len(s.Options.TLS.AutoCertDomains) > 0 {
		s.isTLS = true
		s.Serve(autocert.NewListener(s.Options.TLS.AutoCertDomains...))
		return
	}

	if len(s.Options.TLS.Cert) > 0 {
		s.isTLS = true
		s.Serve(s.enableTLS(s.Listen(), s.Options.TLS.Cert, s.Options.TLS.Key))
		return
	}

	s.Serve(s.Listen())
}

func (s *Server) serveTLS(conn net.Conn) {
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return
	}

	var err error
	if s.Options.ReadTimeout.Duration > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(s.Options.ReadTimeout.Duration))
	}
	err = tlsConn.Handshake()
	if err != nil { // tls handshake error
		return
	}
	if s.Options.ReadTimeout.Duration > 0 {
		_ = conn.SetReadDeadline(time.Time{})
	}
	switch tlsConn.ConnectionState().NegotiatedProtocol {
	case "", "http/1.0", "http/1.1":
		s.serveHTTPConn(tlsConn)
	}
}

func (s *Server) serveHTTPConn(conn net.Conn) {
	s.httpProtocol.ServeConn(context.WithValue(s.baseCtx, CtxKeyConnection, conn), conn)
}

func ListenAndServe(addr string, handler RequestHandler) {
	if handler == nil {
		handler = DefaultMux
	}
	ListenAndServeWithContext(context.Background(), addr, handler)
}

func ListenAndServeWithContext(ctx context.Context, addr string, handler RequestHandler) {
	server := DefaultWithContext(ctx)
	server.Options = defaultServerOption
	server.Options.Addr = addr
	server.Handler = handler
	server.ListenAndServe()
}
