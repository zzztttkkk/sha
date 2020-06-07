package snow

import (
	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/mware"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/redisc"
	"github.com/zzztttkkk/snow/secret"
	"github.com/zzztttkkk/snow/sqls"
)

type Config struct {
	IniFiles   []string
	UserReader mware.UserReader
}

func Init(config *Config) {
	for _, fn := range config.IniFiles {
		ini.Load(fn)
	}

	ini.Init()
	secret.Init()
	output.Init()
	sqls.Init()
	redisc.Init()
	mware.Init(config.UserReader)
}
