package snow

import (
	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/middleware"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/secret"
	"github.com/zzztttkkk/snow/sqls"
)

var userReader middleware.UserFetcher

type InitOption struct {
	IniFiles    []string
	UserFetcher middleware.UserFetcher
}

func Init(opt *InitOption) {
	for _, fn := range opt.IniFiles {
		ini.Load(fn)
	}
	userReader = opt.UserFetcher
	ini.Init()
	secret.Init()
	output.Init()
	sqls.Init()
	middleware.Init(userReader)
}
