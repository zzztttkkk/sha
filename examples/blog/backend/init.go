package backend

import (
	"github.com/zzztttkkk/snow"

	"github.com/zzztttkkk/snow/ini"

	"github.com/zzztttkkk/snow/examples/blog/backend/internal"
)

func Init(conf *ini.Config) {
	internal.LazyExecutor.Execute(snow.Kwargs{"config": conf})
}
