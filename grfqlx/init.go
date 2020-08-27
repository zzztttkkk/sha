package grfqlx

import (
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

var _ShowFmtError bool

func init() {
	internal.Dig.LazyInvoke(
		func(cfg *config.Suna) {
			_ShowFmtError = cfg.Graphql.ShowFormatError
		},
	)
}
