package account

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/examples/blog/backend/models"
	"github.com/zzztttkkk/snow/mware/ctxs"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/sqls"
)

func Unregister(ctx *fasthttp.RequestCtx) {
	uid := ctxs.Uid(ctx)
	if uid < 0 {
		output.StdError(ctx, fasthttp.StatusUnauthorized)
		return
	}
	doUnregister(ctx, uid, string(ctx.FormValue("secret")))
	output.MsgOk(ctx, nil)
}

func doUnregister(ctx context.Context, uid int64, skey string) {
	ctx, committer := sqls.Tx(ctx)
	defer committer()
	models.UserOperator.Delete(ctx, uid, skey)
}
