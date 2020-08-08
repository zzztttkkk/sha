package output

import (
	"github.com/go-errors/errors"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

func init() {
	internal.LazyInvoke(
		func(conf *config.Config) {
			errors.MaxStackDepth = conf.Errors.MaxDepth
			if errors.MaxStackDepth < 1 {
				errors.MaxStackDepth = 20
			}
		},
	)
}
