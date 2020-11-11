package suna

import "github.com/zzztttkkk/suna/internal"

var dig = internal.NewDigContainer()

func Init(conf *Config) {
	dig.Provide(func() *Config { return conf })
	dig.Invoke()
}
