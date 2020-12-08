package rbac

import (
	"context"
	sunainternal "github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/rbac/dao"
	"github.com/zzztttkkk/suna/rbac/internal"
	"github.com/zzztttkkk/suna/rbac/model"
	"log"
)

type Options struct {
	TableNamePrefix  string
	LogReadOperation bool
	Router           Router
	Logger           *log.Logger
}

func Init(options *Options) {
	if len(options.TableNamePrefix) > 0 {
		model.TablenamePrefix = options.TableNamePrefix
	}

	dao.LogReadOperation = options.LogReadOperation

	if options.Logger != nil {
		internal.Logger = options.Logger
	}

	internal.Dig.Provide(func() Router { return options.Router })
	internal.Dig.Append(func(_ internal.DaoOK) { Load(context.Background()) })
	internal.Dig.Invoke()

	sunainternal.RbacInited = true
}
