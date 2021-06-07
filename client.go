package sha

import (
	"context"
	"errors"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type CliOptions struct {
	CliSessionOptions
	MaxAge  int64
	MaxIdle int
	MaxOpen int
	//MaxRedirect
	// <0: no limit on the number of redirects
	// =0: do not redirect
	// >0: limit on the number of redirects
	MaxRedirect         int
	KeepRedirectHistory bool
}

type Cli struct {
	mutex   sync.Mutex
	idling  map[string]chan *CliSession
	using   int64
	Opts    CliOptions
	closing bool
}

var defaultClientSessionPoolOptions = CliOptions{
	defaultCliOptions,
	600, // 10min
	10, 10,
	0,
	false,
}

func NewCli(opt *CliOptions) *Cli {
	cp := &Cli{
		idling: map[string]chan *CliSession{},
	}
	if opt == nil {
		cp.Opts = defaultClientSessionPoolOptions
	} else {
		cp.Opts = *opt
	}
	return cp
}

var ErrClosedCli = errors.New("sha.cli: closed")

func (cp *Cli) _get(ctx context.Context, addr string, isTLS bool) (*CliSession, error) {
	key := fmt.Sprintf("%s:%v", addr, isTLS)

	cp.mutex.Lock()
	if cp.closing {
		cp.mutex.Unlock()
		return nil, ErrClosedCli
	}

	iC := cp.idling[key]
	if iC == nil {
		iC = make(chan *CliSession, cp.Opts.MaxIdle)
		cp.idling[key] = iC
	}
	cp.mutex.Unlock()

	var n time.Duration = 0

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case cli, ok := <-iC:
			if !ok {
				return nil, ErrClosedCli
			}
			return cli, nil
		default:
			if int(atomic.LoadInt64(&cp.using)) < cp.Opts.MaxOpen {
				return NewCliSession(addr, isTLS, &cp.Opts.CliSessionOptions), nil
			}
			n += 2
			if n > 10 {
				n = 2
			}
			time.Sleep(time.Millisecond * n)
		}
	}
}

func (cp *Cli) get(ctx context.Context, addr string, isTLS bool) (*CliSession, error) {
	for {
		session, err := cp._get(ctx, addr, isTLS)
		if err != nil {
			return nil, err
		}
		if cp.Opts.MaxAge > 0 && time.Now().Unix()-session.created > cp.Opts.MaxAge {
			_ = session.Close()
			session = nil
		}
		if session != nil {
			atomic.AddInt64(&cp.using, 1)
			return session, nil
		}
	}
}

func (cp *Cli) put(s *CliSession) {
	if s == nil {
		return
	}

	defer atomic.AddInt64(&cp.using, -1)

	cp.mutex.Lock()

	if cp.closing {
		_ = s.Close()
		cp.mutex.Unlock()
		return
	}

	key := fmt.Sprintf("%s:%v", s.address, s.isTLS)
	iC := cp.idling[key]
	cp.mutex.Unlock()

	for i := 2; i > 0; i-- {
		select {
		case iC <- s:
			return
		default:
			time.Sleep(time.Millisecond * 5)
		}
	}
	_ = s.Close()
}

var ErrMaxRedirect = errors.New("sha.client: reach the redirect limit")

func (cp *Cli) doSend(ctx *RequestCtx, addr string, isTLS bool, onDone func(*CliSession, error), redirectCount int, session *CliSession) {
	if cp.Opts.MaxRedirect > 0 && redirectCount > cp.Opts.MaxRedirect {
		onDone(nil, ErrMaxRedirect)
		return
	}

	reusedSession := session != nil
	var err error
	if !reusedSession {
		session, err = cp.get(ctx, addr, isTLS)
		if err != nil {
			onDone(session, err)
			return
		}
	}
	shouldPutSession := true
	defer func() {
		if reusedSession || !shouldPutSession {
			return
		}
		cp.put(session)
	}()

	if cp.Opts.KeepRedirectHistory {
		ctx.Request.history = append(ctx.Request.history, string(ctx.Request.fl2))
	}

	begin := ctx.Request.time
	err = session.Send(ctx)
	if begin != 0 {
		ctx.Request.time = begin
	}
	if err != nil {
		onDone(session, err)
		return
	}

	res := &ctx.Response
	if res.StatusCode() > 299 && res.StatusCode() < 400 {
		location, _ := res.Header().Get(HeaderLocation)
		if len(location) < 1 {
			onDone(session, err)
			return
		}

		if cp.Opts.MaxRedirect == 0 {
			onDone(session, err)
			return
		}

		u, _ := url.Parse(utils.S(location))
		var _addr string
		var _isTls bool

		if u == nil {
			_addr = addr
			_isTls = isTLS
			ctx.Request.SetPath(location)
		} else {
			if u.Scheme == "" {
				_isTls = isTLS
			} else {
				_isTls = u.Scheme == "https"
			}
			if u.Host == "" {
				_addr = addr
			} else {
				_addr = u.Host
				if !strings.ContainsRune(_addr, ':') {
					if _isTls {
						_addr += ":443"
					} else {
						_addr += ":80"
					}
				}
			}

			ctx.Request.SetPathString(u.Path)
			if len(u.RawQuery) > 0 {
				ctx.Request.fl2 = append(ctx.Request.fl2, '?')
				ctx.Request.fl2 = append(ctx.Request.fl2, u.RawQuery...)
			}
		}

		ctx.Response.reset()

		if _addr == addr && _isTls == isTLS {
			cp.doSend(ctx, addr, isTLS, onDone, redirectCount+1, session)
			return
		}
		cp.put(session)
		shouldPutSession = false
		cp.doSend(ctx, _addr, _isTls, onDone, redirectCount+1, nil)
		return
	}

	onDone(session, nil)
}

func (cp *Cli) Send(ctx *RequestCtx, addr string, onDone func(*CliSession, error)) {
	var isTLS bool
	var ind = strings.Index(addr, "://")
	var protocol string
	if ind > -1 {
		protocol = addr[:ind]
		addr = addr[ind+3:]
	}
	if len(protocol) < 1 {
		protocol = "http"
	}
	switch protocol {
	case "http", "https":
		isTLS = protocol == "https"
		cp.doSend(ctx, addr, isTLS, onDone, 0, nil)
	default:
		onDone(nil, fmt.Errorf("sha.cli: bad protocol: `%s`", protocol))
	}
}

func (cp *Cli) Close() error {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	if cp.closing {
		return nil
	}
	cp.closing = true

	for _, cn := range cp.idling {
		close(cn)
		for v := range cn {
			_ = v.Close()
		}
	}
	return nil
}
