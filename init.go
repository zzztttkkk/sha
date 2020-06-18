package snow

import (
	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/middleware"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/secret"
	"github.com/zzztttkkk/snow/sqls"
)

var iniFiles []string
var userReader middleware.UserFetcher

func AppendIniFile(filename string) {
	iniFiles = append(iniFiles, filename)
}

func SetUserFetcher(fetcher middleware.UserFetcher) {
	userReader = fetcher
}

func Init() {
	for _, fn := range iniFiles {
		ini.Load(fn)
	}
	ini.Init()
	secret.Init()
	output.Init()
	sqls.Init()
	middleware.Init(userReader)
}
