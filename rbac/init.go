package rbac

import (
	"context"
	"github.com/zzztttkkk/suna/rbac/auth"
	"github.com/zzztttkkk/suna/rbac/dao"
	"github.com/zzztttkkk/suna/rbac/internal"
	"github.com/zzztttkkk/suna/rbac/model"
)

type Options struct {
	Authenticator    auth.Authenticator
	TableNamePrefix  string
	LogReadOperation bool
	Router           Router
}

var inited bool

func Init(options *Options) {
	if len(options.TableNamePrefix) > 0 {
		model.TablenamePrefix = options.TableNamePrefix
	}

	dao.LogReadOperation = options.LogReadOperation

	internal.Dig.Provide(func() auth.Authenticator { return options.Authenticator })
	internal.Dig.Provide(func() Router { return options.Router })
	internal.Dig.Append(func(_ internal.DaoOK) { Load(context.Background()) })
	internal.Dig.Invoke()

	inited = true
}
