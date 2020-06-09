package snow

import (
	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/mware"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/secret"
	"github.com/zzztttkkk/snow/sqls"
)

type Config struct {
	IniFiles   []string
	UserReader mware.UidReader
}

func Init(config *Config) {
	for _, fn := range config.IniFiles {
		ini.Load(fn)
	}

	ini.Init()
	secret.Init()
	output.Init()
	sqls.Init()
	mware.Init(config.UserReader)
}
