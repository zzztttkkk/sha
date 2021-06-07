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

// NewCli return a concurrency-safe http client, which holds a session map to reuse connections.
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
				return newCliSession(addr, isTLS, &cp.Opts.CliSessionOptions), nil
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

func (cp *Cli) doSend(ctx *RequestCtx, addr string, isTLS bool, redirectCount int, session *CliSession) error {
	if cp.Opts.MaxRedirect > 0 && redirectCount > cp.Opts.MaxRedirect {
		return ErrMaxRedirect
	}

	reusedSession := session != nil
	var err error
	if !reusedSession {
		session, err = cp.get(ctx, addr, isTLS)
		if err != nil {
			return err
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
		return err
	}

	res := &ctx.Response
	if res.StatusCode() > 299 && res.StatusCode() < 400 {
		if cp.Opts.MaxRedirect == 0 {
			return nil
		}

		location, _ := res.Header().Get(HeaderLocation)
		if len(location) < 1 {
			return nil
		}

		u, _ := url.Parse(utils.S(location))
		var redirectLocationAddr string
		var redirectLocationIsTls bool

		if u == nil {
			redirectLocationAddr = addr
			redirectLocationIsTls = isTLS
			ctx.Request.SetPath(location)
		} else {
			if u.Scheme == "" {
				redirectLocationIsTls = isTLS
			} else {
				redirectLocationIsTls = u.Scheme == "https"
			}
			if u.Host == "" {
				redirectLocationAddr = addr
			} else {
				redirectLocationAddr = u.Host
				if !strings.ContainsRune(redirectLocationAddr, ':') {
					if redirectLocationIsTls {
						redirectLocationAddr += ":443"
					} else {
						redirectLocationAddr += ":80"
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

		if redirectLocationAddr == addr && redirectLocationIsTls == isTLS { // redirect to same host
			return cp.doSend(ctx, addr, isTLS, redirectCount+1, session)
		}

		// redirect to another host
		cp.put(session)
		shouldPutSession = false
		return cp.doSend(ctx, redirectLocationAddr, redirectLocationIsTls, redirectCount+1, nil)
	}
	return nil
}

func (cp *Cli) Send(ctx *RequestCtx, addr string) error {
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
		return cp.doSend(ctx, addr, isTLS, 0, nil)
	default:
		return fmt.Errorf("sha.cli: bad protocol: `%s`", protocol)
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
