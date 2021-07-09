package sha

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zzztttkkk/sha/utils"
)

type CliOptions struct {
	CliConnectionOptions
	MaxAge  int64
	MaxIdle int
	MaxOpen int
	//MaxRedirect
	// <0: no limit on the number of redirects
	// =0: do not redirect
	// >0: limit on the number of redirects
	MaxRedirect         int
	KeepRedirectHistory bool
	EnableCookie        bool
	CookieStoragePath   string
}

type Cli struct {
	mutex   sync.Mutex
	idling  map[string]chan *CliConnection
	using   int64
	Opts    CliOptions
	closing bool
	jar     *CookieJar
}

// NewCli return a concurrency-safe http client, which holds a session map to reuse connections.
func NewCli(opt *CliOptions) *Cli {
	var defaultClientOptions = CliOptions{
		defaultCliOptions,
		600, // 10min
		10, 10,
		0,
		false,
		true,
		"",
	}

	cp := &Cli{
		idling: map[string]chan *CliConnection{},
	}
	if opt == nil {
		cp.Opts = defaultClientOptions
	} else {
		cp.Opts = *opt
		utils.Merge(&cp.Opts, defaultClientOptions)
	}

	if cp.Opts.EnableCookie {
		cp.jar = NewCookieJar()
		if cp.Opts.CookieStoragePath != "" {
			_ = cp.jar.LoadIfExists(cp.Opts.CookieStoragePath)
		}
	}
	return cp
}

var ErrClosedCli = errors.New("sha.cli: closed")

func (cli *Cli) _get(ctx context.Context, addr string, isTLS bool) (*CliConnection, error) {
	key := fmt.Sprintf("%s:%v", addr, isTLS)

	cli.mutex.Lock()

	if cli.closing {
		cli.mutex.Unlock()
		return nil, ErrClosedCli
	}

	iC := cli.idling[key]
	if iC == nil {
		iC = make(chan *CliConnection, cli.Opts.MaxIdle)
		cli.idling[key] = iC
	}
	cli.mutex.Unlock()

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
			if int(atomic.LoadInt64(&cli.using)) < cli.Opts.MaxOpen {
				return newCliConn(addr, isTLS, &cli.Opts.CliConnectionOptions, cli.jar), nil
			}
			n += 2
			if n > 10 {
				n = 2
			}
			time.Sleep(time.Millisecond * n)
		}
	}
}

func (cli *Cli) get(ctx context.Context, addr string, isTLS bool) (*CliConnection, error) {
	for {
		session, err := cli._get(ctx, addr, isTLS)
		if err != nil {
			return nil, err
		}
		if cli.Opts.MaxAge > 0 && time.Now().Unix()-session.created > cli.Opts.MaxAge {
			_ = session.Close()
			session = nil
		}
		if session != nil {
			atomic.AddInt64(&cli.using, 1)
			return session, nil
		}
	}
}

func (cli *Cli) put(s *CliConnection) {
	if s == nil {
		return
	}

	defer atomic.AddInt64(&cli.using, -1)

	cli.mutex.Lock()

	if cli.closing {
		_ = s.Close()
		cli.mutex.Unlock()
		return
	}

	key := fmt.Sprintf("%s:%v", s.address, s.isTLS)
	iC := cli.idling[key]
	cli.mutex.Unlock()

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

func (cli *Cli) doSend(ctx *RequestCtx, addr string, isTLS bool, redirectCount int, session *CliConnection) error {
	if cli.Opts.MaxRedirect > 0 && redirectCount > cli.Opts.MaxRedirect {
		return ErrMaxRedirect
	}

	reusedSession := session != nil
	var err error
	if !reusedSession {
		session, err = cli.get(ctx, addr, isTLS)
		if err != nil {
			return err
		}
	}
	shouldPutSession := true
	defer func() {
		if reusedSession || !shouldPutSession {
			return
		}
		cli.put(session)
	}()

	if cli.Opts.KeepRedirectHistory {
		if session == nil {
			ctx.Request.history = append(ctx.Request.history, fmt.Sprintf("%s%s", addr, utils.S(ctx.Request.fl2)))
		} else {
			ctx.Request.history = append(ctx.Request.history, string(ctx.Request.fl2))
		}
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
		if cli.Opts.MaxRedirect == 0 {
			return nil
		}

		location, _ := res.Header().Get(HeaderLocation)
		if len(location) < 1 {
			return nil
		}

		u, _ := url.Parse(utils.S(location))
		var redirectLocationAddr string
		var redirectLocationIsTLS bool

		if u == nil {
			redirectLocationAddr = addr
			redirectLocationIsTLS = isTLS
			ctx.Request.SetPath(location)
		} else {
			if u.Scheme == "" {
				redirectLocationIsTLS = isTLS
			} else {
				redirectLocationIsTLS = u.Scheme == "https"
			}
			if u.Host == "" {
				redirectLocationAddr = addr
			} else {
				redirectLocationAddr = u.Host
				if !strings.ContainsRune(redirectLocationAddr, ':') {
					if redirectLocationIsTLS {
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

		ctx.Response.reset(cli.Opts.HTTPOptions.BufferPoolSizeLimit)

		if redirectLocationAddr == addr && redirectLocationIsTLS == isTLS { // redirect to same host
			return cli.doSend(ctx, addr, isTLS, redirectCount+1, session)
		}

		// redirect to another host
		cli.put(session)
		shouldPutSession = false
		return cli.doSend(ctx, redirectLocationAddr, redirectLocationIsTLS, redirectCount+1, nil)
	}
	return nil
}

func (cli *Cli) Send(ctx *RequestCtx, addr string) error {
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
		return cli.doSend(ctx, addr, isTLS, 0, nil)
	default:
		return fmt.Errorf("sha.cli: bad protocol: `%s`", protocol)
	}
}

func (cli *Cli) Close() error {
	cli.mutex.Lock()
	defer cli.mutex.Unlock()

	if cli.closing {
		return nil
	}
	cli.closing = true

	for _, cn := range cli.idling {
		close(cn)
		for v := range cn {
			_ = v.Close()
		}
	}

	if cli.jar != nil && cli.Opts.CookieStoragePath != "" {
		return cli.jar.SaveTo(cli.Opts.CookieStoragePath)
	}
	return nil
}
