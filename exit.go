package suna

import (
	"os"
	"os/signal"
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
	signal.Notify(Exit.c, os.Kill, os.Interrupt)

	go func() {
		<-v.c

		for _, call := range v.calls {
			call.fn(call.args...)
		}
		os.Exit(0)
	}()
}

var Exit = &_Exit{c: make(chan os.Signal, 1)}

func init() { Exit.wait() }
