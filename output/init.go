package output

import (
	"github.com/go-errors/errors"

	"github.com/zzztttkkk/snow/ini"
)

func Init() {
	errors.MaxStackDepth = int(ini.GetIntOr("output.max_stack_depth", 20))
}
