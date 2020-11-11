package internal

import (
	"go.uber.org/dig"
)

type _DigInvoke struct {
	Func interface{}
	Opts []dig.InvokeOption
}

type _Dig struct {
	container *dig.Container
	invokes   []_DigInvoke
}

func NewDigContainer(opts ...dig.Option) *_Dig { return &_Dig{container: dig.New(opts...)} }

func (d *_Dig) Provide(constructor interface{}, opts ...dig.ProvideOption) {
	err := d.container.Provide(constructor, opts...)
	if err != nil {
		panic(err)
	}
}

func (d *_Dig) Append(function interface{}, opts ...dig.InvokeOption) {
	d.invokes = append(d.invokes, _DigInvoke{Func: function, Opts: opts})
}

func (d *_Dig) Invoke() {
	for _, v := range d.invokes {
		if err := d.container.Invoke(v.Func, v.Opts...); err != nil {
			panic(err)
		}
	}
}
