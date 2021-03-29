package sha

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"io"
	"net"
	"net/http"
	urllib "net/url"
	"reflect"
	"strings"
	"sync"
	"time"
)

type HTTPSession struct {
	mutex     sync.Mutex
	address   string
	connCount int

	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer

	isTLS bool
	env   Environment
	pool  *RequestCtxPool
}

type HTTPProxy struct {
	Address   string                  `json:"address" toml:"address"`
	AuthFunc  func() (string, string) `json:"-" toml:"-"`
	TLSConfig *tls.Config             `json:"-" toml:"-"`
	IsTLS     bool                    `json:"is_tls" toml:"is-tls"`
}

type Environment struct {
	Header             utils.MultiValueMap `json:"header" toml:"header"`
	Query              utils.MultiValueMap `json:"query" toml:"query"`
	HTTPProxy          HTTPProxy           `json:"http_proxy" toml:"http-proxy"`
	InsecureSkipVerify bool                `json:"insecure_skip_verify" toml:"insecure-skip-verify"`
	TLSConfig          *tls.Config         `json:"-" toml:"-"`
}

func newHTTPSession(address string, pool *RequestCtxPool, env *Environment) *HTTPSession {
	if env == nil {
		env = &Environment{}
	}
	return &HTTPSession{env: *env, address: address, pool: pool}
}

func newHTTPSSession(address string, pool *RequestCtxPool, env *Environment) *HTTPSession {
	if env == nil {
		env = &Environment{}
	}
	return &HTTPSession{env: *env, address: address, isTLS: true, pool: pool}
}

func (s *HTTPSession) Close() error { return s.conn.Close() }

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
		e = parseResponse(ctx, r, make([]byte, 128), &res, s.pool.opt)
		if e != nil {
			return e
		}
		if res.statusCode != StatusOK {
			return fmt.Errorf("sha.clent: bad proxy response, %d %s", res.StatusCode(), res.Phrase())
		}
		if s.isTLS {
			if s.env.TLSConfig == nil {
				s.env.TLSConfig = &tls.Config{}
				if s.env.InsecureSkipVerify {
					s.env.TLSConfig.InsecureSkipVerify = true
				} else {
					s.env.TLSConfig.ServerName = strings.Split(s.address, ":")[0]
				}
			}

			c = tls.Client(proxyC, s.env.TLSConfig)
			r.Reset(c)
			w.Reset(c)
		} else {
			c = proxyC
		}
	} else {
		if s.isTLS {
			c, e = tls.Dial("tcp", s.address, s.env.TLSConfig)
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

func (s *HTTPSession) copyEnv(ctx *RequestCtx) {
	for k, vl := range s.env.Header {
		for _, v := range vl {
			ctx.Request.Header().AppendString(k, v)
		}
	}
}

func (s *HTTPSession) Send(ctx *RequestCtx) error {
	s.copyEnv(ctx)

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
	return parseResponse(ctx, s.r, ctx.readBuf, &ctx.Response, s.pool.opt)
}

var ErrBadValueType = errors.New("sha.client: bad value type")

func isNil(v interface{}) bool {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}
	return rv.IsNil()
}

func buildRequest(req *Request, method, path string, query, header, body interface{}) error {
	req.Reset()
	req.SetPathString(path)
	req.SetMethod(method)

	if !isNil(query) && !req.Query().LoadAny(query) {
		return ErrBadValueType
	}
	if !isNil(header) && !req.Header().LoadAny(header) {
		return ErrBadValueType
	}

	var err error

	if !isNil(body) {
		switch tv := body.(type) {
		case []byte:
			_, _ = req._HTTPPocket.Write(tv)
		case string:
			_, _ = req._HTTPPocket.Write(utils.B(tv))
		case *Form, *utils.Kvs, *Header, urllib.Values, http.Header, utils.MultiValueMap, map[string][]string:
			if !req.bodyForm.LoadAny(body) {
				return err
			}
		case io.Reader:
			var b = make([]byte, 512)
			for {
				l, e := tv.Read(b)
				if l == 0 || e == io.EOF {
					break
				}
				if e != nil {
					return e
				}
				_, _ = req._HTTPPocket.Write(b[:l])
			}
		default:
			if string(req.Header().ContentType()) == MIMEJson {
				err = req.SetJSONBody(body)
			} else {
				err = ErrBadValueType
			}
		}
	}
	return err
}

type ReqAction struct {
	Ext interface{}

	hostName string
	session  *HTTPSession

	method  string
	url     string
	query   interface{}
	header  interface{}
	body    interface{}
	env     *Environment
	onDone  func(res *Response)
	onErr   func(err error)
	err     error
	keepCtx bool
	ctx     *RequestCtx
	pCtx    context.Context

	castTime int64
}

func (action *ReqAction) UseSession(session *HTTPSession, hostName string) *ReqAction {
	action.session = session
	action.hostName = hostName
	return action
}

func (action *ReqAction) Close() error {
	if action.session == nil {
		return nil
	}
	action.ReturnRequestCtx()
	return action.session.Close()
}

func (action *ReqAction) KeepRequestCtx() *ReqAction {
	action.keepCtx = true
	return action
}

func (action *ReqAction) ReturnRequestCtx() *ReqAction {
	if action.ctx != nil {
		action.ctx.ReturnTo(DefaultRequestCtxPool())
		action.ctx = nil
	}
	return action
}

func (action *ReqAction) SetMethod(method string) *ReqAction {
	action.method = method
	return action
}

func (action *ReqAction) SetURL(url string) *ReqAction {
	action.url = url
	return action
}

func (action *ReqAction) SetQuery(query interface{}) *ReqAction {
	action.query = query
	return action
}

func (action *ReqAction) SetHeader(header interface{}) *ReqAction {
	action.header = header
	return action
}

func (action *ReqAction) SetBody(body interface{}) *ReqAction {
	action.body = body
	return action
}

func (action *ReqAction) SetEnv(env *Environment) *ReqAction {
	action.env = env
	return action
}

func (action *ReqAction) OnDone(fn func(res *Response)) *ReqAction {
	action.onDone = fn
	return action
}

func (action *ReqAction) OnError(fn func(error)) *ReqAction {
	action.onErr = fn
	return action
}

func (action *ReqAction) Err() error {
	return action.err
}

func (action *ReqAction) Response() *Response {
	if action.ctx != nil {
		return &action.ctx.Response
	}
	return nil
}

func (action *ReqAction) Send() *ReqAction {
	err := action.do()
	if err != nil {
		action.err = err
		if action.onErr != nil {
			action.onErr(err)
		}
	}
	return action
}

func (action *ReqAction) CastTime() time.Duration { return time.Duration(action.castTime) }

func (action *ReqAction) do() error {
	u, e := urllib.Parse(action.url)
	if e != nil {
		return e
	}

	pool := DefaultRequestCtxPool()
	ctx := pool.Acquire()
	defer pool.Release(ctx)

	ctx.ctx = action.pCtx

	var session = action.session

	var addr = u.Host
	if action.session == nil {
		var addPort = strings.IndexByte(addr, ':') < 0
		if strings.ToLower(u.Scheme) == "https" {
			if addPort {
				addr += ":443"
			}
			session = pool.NewHTTPSSession(addr, action.env)
		} else {
			if addPort {
				addr += "80"
			}
			session = pool.NewHTTPSession(addr, action.env)
		}
		action.session = session
	}

	if len(u.RawQuery) != 0 {
		u.Path = fmt.Sprintf("%s?%s", u.Path, u.RawQuery)
	}
	if e = buildRequest(&ctx.Request, action.method, u.Path, action.query, action.header, action.body); e != nil {
		return e
	}

	if u.User != nil {
		ctx.Request.Header().SetString(HeaderAuthorization, base64.URLEncoding.EncodeToString(utils.B(u.User.String())))
	}

	if len(action.hostName) > 0 {
		ctx.Request.header.SetString(HeaderHost, action.hostName)
	} else {
		ctx.Request.header.SetString(HeaderHost, addr)
	}

	if e = session.Send(ctx); e != nil {
		return e
	}

	action.castTime = ctx.Response.time - ctx.Request.time
	if action.onDone != nil {
		action.onDone(&ctx.Response)
	}
	if action.keepCtx {
		ctx.Keep()
		action.ctx = ctx
	}
	return nil
}

var RequestActionGenerator func() *ReqAction

func init() {
	if RequestActionGenerator == nil {
		RequestActionGenerator = func() *ReqAction { return &ReqAction{} }
	}
}

func newAction(ctx context.Context, url string) *ReqAction {
	action := RequestActionGenerator()
	action.pCtx = ctx
	action.url = url
	return action
}

func Get(ctx context.Context, url string) *ReqAction {
	return newAction(ctx, url).SetMethod(MethodGet)
}

func Post(ctx context.Context, url string) *ReqAction {
	return newAction(ctx, url).SetMethod(MethodPost)
}

func Put(ctx context.Context, url string) *ReqAction {
	return newAction(ctx, url).SetMethod(MethodPut)
}
