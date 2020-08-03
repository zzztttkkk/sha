package output

import (
	"github.com/go-errors/errors"
	"github.com/zzztttkkk/suna/config"
)

func Init(conf *config.Type) {
	errors.MaxStackDepth = conf.Errors.MaxDepth
	if errors.MaxStackDepth < 1 {
		errors.MaxStackDepth = 20
	}
}
