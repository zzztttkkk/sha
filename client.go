package sha

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
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

var clientBufPool = utils.NewBufferPoll(4096)
var headerLineSep = []byte(": ")

var ErrBadResponseContent = errors.New("sha.client: bad response content")
var ErrRequestTimeout = errors.New("sha.client: timeout")
var ErrBadRequest = errors.New("sha.client: bad request")

func readResponse(ctx context.Context, reader *bufio.Reader, res *Response, rctx *RequestCtx) error {
	var lineBuf bytes.Buffer
	var line []byte
	var resetBuf bool
	var bodySize int
	var bodyRemain int
	var bodyReadBuf []byte
	var deadline time.Time
	var handleTimeout bool

	if rctx != nil && rctx.ctx != nil {
		deadline, handleTimeout = ctx.Deadline()
	}

	for {
		if handleTimeout && time.Now().After(deadline) {
			return ErrRequestTimeout
		}

		if res.parseStatus == 2 {
			if bodySize < 1 {
				bodySize = res.Header.ContentLength()
				if bodySize < 1 {
					res.parseStatus++
					return nil
				}
				bodyRemain = bodySize
			}

			if bodyRemain == 0 {
				res.parseStatus++
				return nil
			}

			if len(bodyReadBuf) < 1 {
				bodyReadBuf = make([]byte, 512)
			}

			l, e := reader.Read(bodyReadBuf)
			if e != nil {
				return e
			}

			if l == 0 {
				continue
			}

			if rctx != nil {
				if res.bodyBuf == nil {
					res.bodyBuf = clientBufPool.Get()
					rctx.onReset = append(
						rctx.onReset,
						func(ctx *RequestCtx) { clientBufPool.Put(ctx.Response.bodyBuf) },
					)
				}
				_, _ = res.bodyBuf.Write(bodyReadBuf[:l])
			}

			bodyRemain -= l
			if bodyRemain == 0 {
				res.parseStatus++
				return nil
			}
			if bodyRemain < 0 {
				return ErrBadResponseContent
			}
			continue
		}

		linePart, isPrefix, err := reader.ReadLine()
		if err != nil {
			return err
		}

		if resetBuf {
			lineBuf.Reset()
		}

		if isPrefix {
			lineBuf.Write(linePart)
			continue
		}

		if lineBuf.Len() > 0 {
			lineBuf.Write(linePart)
			line = lineBuf.Bytes()
			resetBuf = true
		} else {
			line = linePart
		}

		switch res.parseStatus {
		case 0:
			fs := 0
			status := make([]byte, 0, 10)
			ind := 0
			var b byte
			for ind, b = range line {
				if b == ' ' {
					fs++
					if fs == 2 {
						break
					}
					continue
				}
				switch fs {
				case 0:
					res.version = append(res.version, b)
				case 1:
					status = append(status, b)
				}
			}

			if fs != 2 {
				return ErrBadResponseContent
			}

			res.setPhrase(line[ind:])
			s, e := strconv.ParseInt(utils.S(status), 10, 64)
			if e != nil || s < 0 || s > math.MaxInt32 {
				return ErrBadResponseContent
			}
			res.statusCode = int(s)

			res.parseStatus++
		case 1:
			if len(line) < 1 {
				res.parseStatus++
				continue
			}
			ind := bytes.Index(line, headerLineSep)
			if ind < 1 {
				return ErrBadResponseContent
			}
			res.Header.Set(utils.S(line[:ind]), line[ind+2:])
		}
	}
}

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
		e = readResponse(ctx, r, &res, nil)
		if e != nil {
			return e
		}
		if res.statusCode != StatusOK {
			return fmt.Errorf("sha.clent: bad proxy response, %d %s", res.statusCode, res.phrase)
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
			ctx.Request.Header.AppendString(k, v)
		}
	}
}

const _clientBufKey = "sha.client.buf"

func (s *Connection) Send(ctx *RequestCtx) error {
	if err := s.openConn(ctx); err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	ctx.isTLS = s.isTLS
	s.copyEnv(ctx)

	sendBuf := clientBufPool.Get()
	defer clientBufPool.Put(sendBuf)

	req := &ctx.Request

	_, _ = sendBuf.Write(req.Method)
	_ = sendBuf.WriteByte(' ')
	_, _ = sendBuf.Write(req.Path)
	querySize := req.query.Size()
	if querySize != 0 {
		_ = sendBuf.WriteByte('?')
		ind := 0
		req.query.EachItem(func(item *utils.KvItem) bool {
			utils.EncodeURIComponent(item.Key, &sendBuf.Data)
			_ = sendBuf.WriteByte('=')
			utils.EncodeURIComponent(item.Val, &sendBuf.Data)
			ind++
			if ind < querySize {
				_ = sendBuf.WriteByte('&')
			}
			return true
		})
	}
	sendBuf.WriteString(" HTTP/1.1")
	sendBuf.WriteString("\r\n")

	bodyFormSize := req.body.Size()
	if bodyFormSize > 0 {
		if req.bodyBufferPtr != nil {
			return ErrBadRequest
		}
		buf := clientBufPool.Get()
		ctx.ud.Set(_clientBufKey, buf)
		ctx.onReset = append(
			ctx.onReset,
			func(_ctx *RequestCtx) {
				v := _ctx.ud.Get(_clientBufKey)
				if v != nil {
					clientBufPool.Put(v.(*utils.Buf))
				}
			},
		)

		req.bodyBufferPtr = &buf.Data
		ind := 0
		req.body.EachItem(func(item *utils.KvItem) bool {
			utils.EncodeURIComponent(item.Key, &buf.Data)
			_ = buf.WriteByte('=')
			utils.EncodeURIComponent(item.Val, &buf.Data)
			ind++
			if ind < bodyFormSize {
				_ = sendBuf.WriteByte('&')
			}
			return true
		})
	}

	var bodySize int
	if req.bodyBufferPtr != nil {
		bodySize = len(*req.bodyBufferPtr)
		if bodySize > 0 {
			req.Header.SetContentLength(int64(bodySize))
		}
	}

	req.Header.EachItem(func(item *utils.KvItem) bool {
		_, _ = sendBuf.Write(item.Key)
		sendBuf.WriteString(": ")
		_, _ = sendBuf.Write(item.Val)
		sendBuf.WriteString("\r\n")
		return true
	})
	sendBuf.WriteString("\r\n")
	if bodySize > 0 {
		_, _ = sendBuf.Write(*req.bodyBufferPtr)
	}

	_, err := s.w.Write(sendBuf.Data)
	if err != nil {
		return err
	}
	err = s.w.Flush()
	if err != nil {
		return err
	}
	return readResponse(ctx, s.r, &ctx.Response, ctx)
}
