package internal

import (
	"go.uber.org/dig"
)

type _DigInvoke struct {
	Func interface{}
	Opts []dig.InvokeOption
}

type _Dig struct {
	Container *dig.Container
	Invokes   []_DigInvoke
}

func NewDigContainer(opts ...dig.Option) *_Dig { return &_Dig{Container: dig.New(opts...)} }

func (d *_Dig) Provide(constructor interface{}, opts ...dig.ProvideOption) {
	err := d.Container.Provide(constructor, opts...)
	if err != nil {
		panic(err)
	}
}

func (d *_Dig) LazyInvoke(function interface{}, opts ...dig.InvokeOption) {
	d.Invokes = append(d.Invokes, _DigInvoke{Func: function, Opts: opts})
}

func (d *_Dig) Invoke() {
	for _, v := range d.Invokes {
		if err := d.Container.Invoke(v.Func, v.Opts...); err != nil {
			panic(err)
		}
	}
}

var Dig = NewDigContainer()
