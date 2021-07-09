package sha

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"net"
	"strings"
	"time"
)

type CliConnection struct {
	address string
	host    string

	created int64

	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer

	isTLS bool
	opt   *CliConnectionOptions

	jar *CookieJar
}

type HTTPProxy struct {
	Address   string
	AuthFunc  func() (string, string)
	TLSConfig *tls.Config
	IsTLS     bool
}

type CliConnectionOptions struct {
	HTTPOptions          *HTTPOptions
	HTTPProxy            HTTPProxy
	InsecureSkipVerify   bool
	TLSConfig            *tls.Config
	BeforeSendRequest    []func(ctx *RequestCtx, host string) error
	AfterReceiveResponse []func(ctx *RequestCtx, err error)
}

func (o *CliConnectionOptions) h2tpOpts() *HTTPOptions {
	if o.HTTPOptions != nil {
		return o.HTTPOptions
	}
	return &defaultHTTPOption
}

var defaultCliOptions CliConnectionOptions

func newCliConn(address string, isTLS bool, opt *CliConnectionOptions, jar *CookieJar) *CliConnection {
	if opt == nil {
		opt = &defaultCliOptions
	}

	s := &CliConnection{opt: opt, isTLS: isTLS, created: time.Now().Unix(), jar: jar}
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
	return s
}

func (conn *CliConnection) Close() error {
	if conn.conn == nil {
		return nil
	}

	e := conn.conn.Close()
	conn.conn = nil
	return e
}

func (conn *CliConnection) Reconnect() {
	if conn.conn == nil {
		return
	}

	_ = conn.conn.Close()
	conn.conn = nil
	if conn.r != nil {
		conn.r.Reset(nil)
	}
	if conn.w != nil {
		conn.w.Reset(nil)
	}
}

func (conn *CliConnection) openConn(ctx context.Context) error {
	if conn.conn != nil {
		return nil
	}

	var c net.Conn
	var e error
	var w = conn.w
	var r = conn.r

	proxy := &conn.opt.HTTPProxy
	if len(proxy.Address) > 0 {
		var buf bytes.Buffer
		buf.WriteString("CONNECT ")
		buf.WriteString(conn.address)
		buf.WriteByte(' ')
		buf.WriteString("HTTP/1.1\r\n")
		buf.WriteString("Host: ")
		buf.WriteString(conn.address)
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
		e = parseResponse(ctx, r, make([]byte, 128), &res, conn.opt.h2tpOpts())
		if e != nil {
			return e
		}
		if res.statusCode != StatusOK {
			return fmt.Errorf("sha.clent: bad proxy response, %d %s", res.StatusCode(), res.Phrase())
		}
		if conn.isTLS {
			if conn.opt.TLSConfig == nil {
				conn.opt.TLSConfig = &tls.Config{}
				if conn.opt.InsecureSkipVerify {
					conn.opt.TLSConfig.InsecureSkipVerify = true
				} else {
					conn.opt.TLSConfig.ServerName = strings.Split(conn.address, ":")[0]
				}
			}

			c = tls.Client(proxyC, conn.opt.TLSConfig)
		} else {
			c = proxyC
		}
	} else {
		if conn.isTLS {
			c, e = tls.Dial("tcp", conn.address, conn.opt.TLSConfig)
		} else {
			c, e = net.Dial("tcp", conn.address)
		}
	}

	if e != nil {
		return e
	}
	conn.conn = c
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
	conn.r = r
	conn.w = w
	return nil
}

func (conn *CliConnection) Send(ctx *RequestCtx) error {
	for _, fn := range conn.opt.BeforeSendRequest {
		if err := fn(ctx, conn.host); err != nil {
			return err
		}
	}

	if conn.jar != nil {
		conn.jar.toRCtx(ctx, conn.host)
	}

	if ctx.ctx == nil {
		ctx.ctx = context.Background()
	}

	if ctx.readBuf == nil {
		ctx.readBuf = make([]byte, 512)
	}

	if err := conn.openConn(ctx); err != nil {
		return err
	}

	if err := sendRequest(conn.w, &ctx.Request); err != nil {
		return err
	}
	err := parseResponse(ctx, conn.r, ctx.readBuf, &ctx.Response, conn.opt.h2tpOpts())
	if conn.jar != nil && err == nil {
		for _, v := range ctx.Response.Header().GetAll(HeaderSetCookie) {
			_ = conn.jar.Update(conn.host, utils.S(v))
		}
	}
	for _, fn := range conn.opt.AfterReceiveResponse {
		fn(ctx, err)
	}
	return err
}

func (conn *CliConnection) Conn() net.Conn { return conn.conn }
