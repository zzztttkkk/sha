package rbac

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/ini"
	"github.com/zzztttkkk/suna/utils"
)

var tablePrefix string
var lazier = utils.NewLazyExecutor()
var initPriority = utils.NewPriority(1)
var permTablePriority = initPriority.Incr()
var rbacPriority = permTablePriority.Incr().Incr()
var loader = utils.NewLoader()

var wrapRCtx func(ctx *fasthttp.RequestCtx) context.Context
var getUserFromCtx func(ctx context.Context) User
var getUserFromRCtx func(ctx *fasthttp.RequestCtx) User

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

var conf *ini.Ini

func Init(
	confV *ini.Ini,
	wrapCtxFn func(ctx *fasthttp.RequestCtx) context.Context,
	getUserFn func(ctx context.Context) User,
	getUserRctxFn func(ctx *fasthttp.RequestCtx) User,
) {
	conf = confV
	wrapRCtx = wrapCtxFn
	getUserFromCtx = getUserFn
	getUserFromRCtx = getUserRctxFn

	l, _ := conf.SqlClients()
	if l != nil {
		lazier.Execute(nil)
	}
}
