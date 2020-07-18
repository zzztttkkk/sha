package snow

import (
	"github.com/zzztttkkk/snow/sqls"

	"github.com/zzztttkkk/snow/secret"

	"github.com/zzztttkkk/snow/output"

	"github.com/zzztttkkk/snow/internal"

	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/middleware"
)

type InitOption struct {
	IniFiles []string
	Auther   middleware.Auther
}

var config = ini.New()

func Init(opt *InitOption) *ini.Config {
	internal.Provide(
		func() *ini.Config {
			for _, fn := range opt.IniFiles {
				config.Load(fn)
			}
			config.Done()
			config.Print()
			return config
		},
	)
	internal.Provide(func() middleware.Auther { return opt.Auther })

	internal.Invoke(middleware.Init)
	internal.Invoke(output.Init)
	internal.Invoke(secret.Init)
	internal.Invoke(sqls.Init)

	return config
}
