package sha

import (
	"context"
	"fmt"
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
	MaxRedirect int
}

type Cli struct {
	mutex  sync.Mutex
	idling map[string]chan *CliSession
	using  int64
	opts   CliOptions
}

var defaultClientSessionPoolOptions = CliOptions{
	defaultCliOptions,
	600, // 10min
	10, 10,
	0,
}

func NewCli(opt *CliOptions) *Cli {
	cp := &Cli{
		idling: map[string]chan *CliSession{},
	}
	if opt == nil {
		cp.opts = defaultClientSessionPoolOptions
	} else {
		cp.opts = *opt
	}
	return cp
}

func (cp *Cli) _get(ctx context.Context, addr string, isTLS bool) (*CliSession, error) {
	key := fmt.Sprintf("%s:%v", addr, isTLS)

	cp.mutex.Lock()
	iC := cp.idling[key]
	if iC == nil {
		iC = make(chan *CliSession, cp.opts.MaxIdle)
		cp.idling[key] = iC
	}
	cp.mutex.Unlock()

	for {
		select {
		case <-ctx.Done():
			return nil, ErrCanceled
		case cli := <-iC:
			return cli, nil
		default:
			if int(atomic.LoadInt64(&cp.using)) < cp.opts.MaxOpen {
				return NewCliSession(addr, isTLS, &cp.opts.CliSessionOptions), nil
			}
			time.Sleep(time.Millisecond)
		}
	}
}

func (cp *Cli) get(ctx context.Context, addr string, isTLS bool) (*CliSession, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ErrCanceled
		default:
			cli, err := cp._get(ctx, addr, isTLS)
			if err != nil {
				return nil, err
			}
			if cp.opts.MaxAge > 0 && time.Now().Unix()-cli.created > cp.opts.MaxAge {
				_ = cli.Close()
				cli = nil
			}
			if cli != nil {
				atomic.AddInt64(&cp.using, 1)
				return cli, nil
			}
		}
	}
}

func (cp *Cli) put(s *CliSession) {
	cp.mutex.Lock()
	key := fmt.Sprintf("%s:%v", s.address, s.isTLS)
	iC := cp.idling[key]
	defer cp.mutex.Unlock()

	atomic.AddInt64(&cp.using, -1)

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

func (cp *Cli) Send(ctx *RequestCtx, addr string, isTLS bool, onDone func(*CliSession, error)) {
	cli, err := cp.get(ctx, addr, isTLS)
	if err != nil {
		onDone(cli, err)
		return
	}
	defer cp.put(cli)

	onDone(cli, cli.Send(ctx))
}
