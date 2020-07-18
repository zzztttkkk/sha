package account

import (
	"context"

	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/snow/examples/blog/backend/models"
	"github.com/zzztttkkk/snow/middleware/ctxs"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/sqls"
)

func Unregister(ctx *fasthttp.RequestCtx) {
	user := ctxs.User(ctx)
	if user == nil {
		output.StdError(ctx, fasthttp.StatusUnauthorized)
		return
	}
	doUnregister(ctx, user.GetId(), string(ctx.FormValue("secret")))
	output.MsgOK(ctx, nil)
}

func doUnregister(ctx context.Context, uid int64, skey string) {
	ctx, committer := sqls.Tx(ctx)
	defer committer()
	models.UserOperator.Delete(ctx, uid, skey)
}
