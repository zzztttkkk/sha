package output

import (
	"github.com/go-errors/errors"

	"github.com/zzztttkkk/snow/ini"
)

func Init(conf *ini.Config) {
	errors.MaxStackDepth = int(conf.GetIntOr("output.max_stack_depth", 20))
}
