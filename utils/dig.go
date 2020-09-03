package utils

import "go.uber.org/dig"

type DigInvoke struct {
	Func interface{}
	Opts []dig.InvokeOption
}

type Dig struct {
	Container *dig.Container
	Invokes   []DigInvoke
}

func NewDig(opts ...dig.Option) *Dig { return &Dig{Container: dig.New(opts...)} }

func (d *Dig) Provide(constructor interface{}, opts ...dig.ProvideOption) {
	err := d.Container.Provide(constructor, opts...)
	if err != nil {
		panic(err)
	}
}

func (d *Dig) LazyInvoke(function interface{}, opts ...dig.InvokeOption) {
	d.Invokes = append(d.Invokes, DigInvoke{Func: function, Opts: opts})
}

func (d *Dig) Invoke() {
	for _, v := range d.Invokes {
		if err := d.Container.Invoke(v.Func, v.Opts...); err != nil {
			panic(err)
		}
	}
}
