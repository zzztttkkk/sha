package output

import (
	"github.com/go-errors/errors"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

func init() {
	internal.Dig.LazyInvoke(
		func(conf *config.Suna) {
			errors.MaxStackDepth = conf.Errors.MaxDepth
			if errors.MaxStackDepth < 1 {
				errors.MaxStackDepth = 20
			}
		},
	)
}
