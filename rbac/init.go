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

var tablePrefix string
var lazier = utils.NewLazyExecutor()
var initPriority = utils.NewPriority(1)
var permTablePriority = initPriority.Incr()
var rbacPriority = permTablePriority.Incr().Incr()
var loader = utils.NewLoader()

var wrapRCtx func(ctx *fasthttp.RequestCtx) context.Context
var getUserFromCtx func(ctx context.Context) auth.User
var getUserFromRCtx func(ctx *fasthttp.RequestCtx) auth.User

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

var cfg *config.Type

func Init(
	confV *config.Type,
	di *internal.RbacDi,
) {
	cfg = confV
	wrapRCtx = di.WrapCtx
	getUserFromCtx = di.GetUserFromCtx
	getUserFromRCtx = di.GetUserFromRCtx

	l := cfg.SqlLeader()
	if l == nil {
		log.Println("suna.rbac: init error")
		return
	}
	lazier.Execute(nil)
}
