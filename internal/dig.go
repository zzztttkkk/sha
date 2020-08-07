package internal

import (
	"go.uber.org/dig"
)

var container = dig.New()

func Provide(constructor interface{}, opts ...dig.ProvideOption) {
	err := container.Provide(constructor, opts...)
	if err != nil {
		panic(err)
	}
}

type _InvokeArgs struct {
	fn   interface{}
	opts []dig.InvokeOption
}

var _invokes []_InvokeArgs

func LazyInvoke(function interface{}, opts ...dig.InvokeOption) {
	_invokes = append(_invokes, _InvokeArgs{fn: function, opts: opts})
}

func Invoke() {
	for _, v := range _invokes {
		if err := container.Invoke(v.fn, v.opts...); err != nil {
			panic(err)
		}
	}
}
