package output

import (
	"github.com/go-errors/errors"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

var _JsonPCallbackParams string

func init() {
	internal.Dig.LazyInvoke(
		func(conf *config.Suna) {
			errors.MaxStackDepth = conf.Output.ErrorMaxDepth
			if errors.MaxStackDepth < 1 {
				errors.MaxStackDepth = 20
			}
			_JsonPCallbackParams = conf.Output.JsonPCallbackParam
		},
	)
}
