package suna

import (
	"os"
	"os/signal"
	"syscall"
)

type _Call struct {
	fn   func(...interface{})
	args []interface{}
}

type _Exit struct {
	calls []*_Call
	c     chan os.Signal
}

func (v *_Exit) Register(fn func(...interface{}), args ...interface{}) {
	v.calls = append(v.calls, &_Call{fn: fn, args: args})
}

func (v *_Exit) wait() {
	go func() {
		<-v.c

		for _, call := range v.calls {
			call.fn(call.args...)
		}
		os.Exit(0)
	}()

	signal.Notify(
		v.c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		os.Interrupt,
	)
}

var Exit = &_Exit{c: make(chan os.Signal, 1)}

func init() { Exit.wait() }
