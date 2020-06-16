package snow

import (
	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/mware"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/secret"
	"github.com/zzztttkkk/snow/sqls"
)

var iniFiles []string
var userReader mware.UserFetcher

func AppendIniFile(filename string) {
	iniFiles = append(iniFiles, filename)
}

func SetUserFetcher(fetcher mware.UserFetcher) {
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
	mware.Init(userReader)
}
