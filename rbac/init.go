package rbac

import (
	"context"
	shainternal "github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/rbac/dao"
	"github.com/zzztttkkk/sha/rbac/internal"
	"github.com/zzztttkkk/sha/rbac/model"
	"github.com/zzztttkkk/sha/validator"
	"log"
	"regexp"
)

type Options struct {
	TableNamePrefix  string
	Logger           *log.Logger
	LogReadOperation bool
}

var nameRegexp = regexp.MustCompile("^[a-zA-Z0-9_]+(\\.[a-zA-Z0-9_]+)*$")

var gAdapter CtxAdapter

func Init(router Router, adapter CtxAdapter, options *Options) {
	validator.RegisterRegexp("rbacname", nameRegexp)

	if options == nil {
		options = &Options{}
	}

	gAdapter = adapter

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
