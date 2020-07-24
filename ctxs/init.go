package ctxs

import "github.com/zzztttkkk/suna/ini"

var config *ini.Ini

func Init(conf *ini.Ini) {
	config = conf
	_initSession()
}
