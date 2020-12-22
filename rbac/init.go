package rbac

import (
	"context"
	shainternal "github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/rbac/dao"
	"github.com/zzztttkkk/sha/rbac/internal"
	"github.com/zzztttkkk/sha/rbac/model"
	"log"
)

type Options struct {
	TableNamePrefix  string
	LogReadOperation bool
	Logger           *log.Logger
}

func Init(router Router, options *Options) {
	if options == nil {
		options = &Options{}
	}

	if len(options.TableNamePrefix) > 0 {
		model.TablenamePrefix = options.TableNamePrefix
	}

	dao.LogReadOperation = options.LogReadOperation

	if options.Logger != nil {
		internal.Logger = options.Logger
	}

	internal.Dig.Provide(func() Router { return router })
	internal.Dig.Append(func(_ internal.DaoOK) { Load(context.Background()) })
	internal.Dig.Invoke()

	shainternal.RbacInited = true
}
