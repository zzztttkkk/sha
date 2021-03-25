package sha

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Connection struct {
	mutex     sync.Mutex
	address   string
	connCount int

	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer

	isTLS     bool
	tlsConfig *tls.Config
	env       Environment
}

type MultiValueMap map[string][]string

type HTTPProxy struct {
	Address   string                  `json:"address" toml:"address"`
	AuthFunc  func() (string, string) `json:"-" toml:"-"`
	TLSConfig *tls.Config             `json:"-" toml:"-"`
	IsTLS     bool                    `json:"is_tls" toml:"is-tls"`
}

type Environment struct {
	Header             MultiValueMap `json:"header" toml:"header"`
	HTTPProxy          HTTPProxy     `json:"http_proxy" toml:"http-proxy"`
	InsecureSkipVerify bool          `json:"insecure_skip_verify" toml:"insecure-skip-verify"`
}

func NewConnection(address string, env *Environment) *Connection {
	if env == nil {
		env = &Environment{}
	}
	return &Connection{env: *env, address: address}
}

func NewTLSConnection(address string, tlsConf *tls.Config, env *Environment) *Connection {
	if env == nil {
		env = &Environment{}
	}
	return &Connection{env: *env, tlsConfig: tlsConf, address: address, isTLS: true}
}

func (s *Connection) Close() error { return s.conn.Close() }

func (s *Connection) Reconnect() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.conn == nil {
		return
	}

	_ = s.conn.Close()
	s.conn = nil
	if s.r != nil {
		s.r.Reset(nil)
	}
	if s.w != nil {
		s.w.Reset(nil)
	}
}

func (s *Connection) openConn(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.conn != nil {
		return nil
	}

	var c net.Conn
	var e error
	var w = s.w
	var r = s.r

	proxy := &s.env.HTTPProxy
	if len(proxy.Address) > 0 {
		var buf bytes.Buffer
		buf.WriteString("CONNECT ")
		buf.WriteString(s.address)
		buf.WriteByte(' ')
		buf.WriteString("HTTP/1.1\r\n")
		buf.WriteString("Host: ")
		buf.WriteString(s.address)
		buf.WriteString("\r\n")

		if proxy.AuthFunc != nil {
			authType, realm := proxy.AuthFunc()
			buf.WriteString("Proxy-Authenticate: ")
			buf.WriteString(authType)
			buf.WriteByte(' ')
			buf.WriteString(realm)
			buf.WriteString("\r\n")
		}
		buf.WriteString("\r\n")

		var proxyC net.Conn
		if proxy.IsTLS {
			proxyC, e = tls.Dial("tcp", proxy.Address, proxy.TLSConfig)
		} else {
			proxyC, e = net.Dial("tcp", proxy.Address)
		}
		if e != nil {
			return e
		}

		if w == nil {
			w = bufio.NewWriter(proxyC)
		} else {
			w.Reset(proxyC)
		}
		_, e = w.Write(buf.Bytes())
		if e != nil {
			return e
		}
		e = w.Flush()
		if e != nil {
			return e
		}

		var res Response
		if r == nil {
			r = bufio.NewReader(proxyC)
		} else {
			r.Reset(proxyC)
		}

		e = parseResponse(ctx, r, make([]byte, 128), &res)
		if e != nil {
			return e
		}
		if res.statusCode != StatusOK {
			return fmt.Errorf("sha.clent: bad proxy response, %d %s", res.StatusCode(), res.Phrase())
		}
		if s.isTLS {
			if s.tlsConfig == nil {
				s.tlsConfig = &tls.Config{}
				if s.env.InsecureSkipVerify {
					s.tlsConfig.InsecureSkipVerify = true
				} else {
					s.tlsConfig.ServerName = strings.Split(s.address, ":")[0]
				}
			}

			c = tls.Client(proxyC, s.tlsConfig)
			r.Reset(c)
			w.Reset(c)
		} else {
			c = proxyC
		}
	} else {
		if s.isTLS {
			c, e = tls.Dial("tcp", s.address, s.tlsConfig)
		} else {
			c, e = net.Dial("tcp", s.address)
		}
	}

	if e != nil {
		return e
	}
	s.conn = c
	if r == nil {
		r = bufio.NewReader(c)
	}
	if w == nil {
		w = bufio.NewWriter(c)
	}
	s.r = r
	s.w = w
	return nil
}

func (s *Connection) copyEnv(ctx *RequestCtx) {
	for k, vl := range s.env.Header {
		for _, v := range vl {
			ctx.Request.Header().AppendString(k, v)
		}
	}
}

func (s *Connection) Send(ctx *RequestCtx) error {
	s.copyEnv(ctx)

	if ctx.ctx == nil {
		ctx.ctx = context.Background()
	}

	if ctx.readBuf == nil {
		ctx.readBuf = make([]byte, 512)
	}

	if err := s.openConn(ctx); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := sendRequest(s.w, &ctx.Request); err != nil {
		return err
	}
	return parseResponse(ctx, s.r, ctx.readBuf, &ctx.Response)
}
