package internal

import "go.uber.org/dig"

var container = dig.New()

func Provide(constructor interface{}, opts ...dig.ProvideOption) {
	err := container.Provide(constructor, opts...)
	if err != nil {
		panic(err)
	}
}

func Invoke(function interface{}, opts ...dig.InvokeOption) {
	err := container.Invoke(function, opts...)
	if err != nil {
		panic(err)
	}
}
