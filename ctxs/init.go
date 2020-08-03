package ctxs

import (
	"github.com/zzztttkkk/suna/config"
)

var cfg *config.Type

func Init(conf *config.Type) {
	cfg = conf
	_initSession()
}
