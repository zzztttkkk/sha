package rbac

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/utils"
	"log"
)

var tablePrefix = "rbac_"
var lazier = utils.NewLazyExecutor()
var initPriority = utils.NewPriority(1)
var permTablePriority = initPriority.Incr()
var rbacPriority = permTablePriority.Incr().Incr()
var loader = utils.NewLoader()

var wrapRCtx func(ctx *fasthttp.RequestCtx) context.Context
var getUserFromCtx func(ctx context.Context) auth.User

type modifyType int

const (
	Del = modifyType(-1)
	Add = modifyType(1)
)

func (v modifyType) String() string {
	switch v {
	case Add:
		return "add"
	case Del:
		return "del"
	default:
		return ""
	}
}

func Loader() *utils.Loader {
	lazier.RegisterWithPriority(
		func(kwargs utils.Kwargs) { Load(context.Background()) },
		rbacPriority,
	)
	return loader
}

var cfg *config.Config

func init() {
	internal.LazyInvoke(
		func(confV *config.Config, di *internal.RbacDi) {
			cfg = confV
			wrapRCtx = di.WrapCtx.(func(ctx *fasthttp.RequestCtx) context.Context)
			getUserFromCtx = di.GetUserFromCtx.(func(ctx context.Context) auth.User)
			tablePrefix = confV.Rbac.TablenamePrefix

			l := cfg.SqlLeader()
			if l == nil {
				log.Println("suna.rbac: init error")
				return
			}
			lazier.Execute(nil)
		},
	)
}
