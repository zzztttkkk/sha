package snow

import (
	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/mware"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/secret"
	"github.com/zzztttkkk/snow/sqls"
)

var iniFiles []string
var userReader mware.UserReader

func AppendIniFile(filename string) {
	iniFiles = append(iniFiles, filename)
}

func SetUserReader(reader mware.UserReader) {
	userReader = reader
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
