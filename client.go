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

type HTTPSession struct {
	mutex     sync.Mutex
	address   string
	host      string
	connCount int

	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer

	isTLS   bool
	opt     *SessionOptions
	httpOpt *HTTPOptions
}

type HTTPProxy struct {
	Address   string
	AuthFunc  func() (string, string)
	TLSConfig *tls.Config
	IsTLS     bool
}

type SessionOptions struct {
	HTTPOptions          *HTTPOptions
	HTTPProxy            HTTPProxy
	InsecureSkipVerify   bool
	TLSConfig            *tls.Config
	RequestCtxPreChecker func(ctx *RequestCtx)
}

var defaultSessionOptions SessionOptions

func NewHTTPSession(address string, isTLS bool, opt *SessionOptions) *HTTPSession {
	if opt == nil {
		opt = &defaultSessionOptions
	}

	s := &HTTPSession{opt: opt, isTLS: isTLS}
	ind := strings.IndexRune(address, ':')
	if ind > -1 {
		s.host = address[:ind]
		port := address[ind:]
		if isTLS && port != ":443" {
			s.host = address
		}
		if !isTLS && port != ":80" {
			s.host = address
		}
	} else {
		s.host = address
		if isTLS {
			address += ":443"
		} else {
			address += ":80"
		}
	}
	s.address = address

	if s.opt.HTTPOptions != nil {
		s.httpOpt = s.opt.HTTPOptions
	} else {
		s.httpOpt = &defaultHTTPOption
	}
	return s
}

func (s *HTTPSession) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.conn == nil {
		return nil
	}
	return s.conn.Close()
}

func (s *HTTPSession) Reconnect() {
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

func (s *HTTPSession) OpenConn(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.conn != nil {
		return nil
	}

	var c net.Conn
	var e error
	var w = s.w
	var r = s.r

	proxy := &s.opt.HTTPProxy
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
		e = parseResponse(ctx, r, make([]byte, 128), &res, s.httpOpt)
		if e != nil {
			return e
		}
		if res.statusCode != StatusOK {
			return fmt.Errorf("sha.clent: bad proxy response, %d %s", res.StatusCode(), res.Phrase())
		}
		if s.isTLS {
			if s.opt.TLSConfig == nil {
				s.opt.TLSConfig = &tls.Config{}
				if s.opt.InsecureSkipVerify {
					s.opt.TLSConfig.InsecureSkipVerify = true
				} else {
					s.opt.TLSConfig.ServerName = strings.Split(s.address, ":")[0]
				}
			}

			c = tls.Client(proxyC, s.opt.TLSConfig)
		} else {
			c = proxyC
		}
	} else {
		if s.isTLS {
			c, e = tls.Dial("tcp", s.address, s.opt.TLSConfig)
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
	} else {
		r.Reset(c)
	}
	if w == nil {
		w = bufio.NewWriter(c)
	} else {
		w.Reset(c)
	}
	s.r = r
	s.w = w
	return nil
}

func (s *HTTPSession) Send(ctx *RequestCtx) error {
	if s.opt.RequestCtxPreChecker != nil {
		s.opt.RequestCtxPreChecker(ctx)
	}

	if ctx.ctx == nil {
		ctx.ctx = context.Background()
	}

	if ctx.readBuf == nil {
		ctx.readBuf = make([]byte, 512)
	}

	if err := s.OpenConn(ctx); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := sendRequest(s.w, &ctx.Request); err != nil {
		return err
	}
	return parseResponse(ctx, s.r, ctx.readBuf, &ctx.Response, s.httpOpt)
}
