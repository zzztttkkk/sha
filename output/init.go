package output

import (
	"github.com/go-errors/errors"

	"github.com/zzztttkkk/suna/ini"
)

func Init(conf *ini.Ini) {
	errors.MaxStackDepth = int(conf.GetIntOr("output.error.max_depth", 20))
}
