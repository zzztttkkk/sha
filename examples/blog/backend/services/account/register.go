package account

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/examples/blog/backend/models"
	"github.com/zzztttkkk/snow/mware/ctxs"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/sqls"
	"github.com/zzztttkkk/snow/utils"
	"github.com/zzztttkkk/snow/validator"
)

type RegisterForm struct {
	Name     []byte `validator:":F<username>;L<3-20>"`
	Password []byte `validator:":R<password>"`
}

func Register(ctx *fasthttp.RequestCtx) {
	session := ctxs.Session(ctx)
	if !session.CaptchaVerify(ctx) {
		output.StdError(ctx, fasthttp.StatusBadRequest)
		return
	}

	form := RegisterForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	uid, skey := doRegister(ctx, form.Name, form.Password)
	output.MsgOk(ctx, output.M{"uid": uid, "secret": utils.B2s(skey)})
}

func doRegister(ctx context.Context, name, password []byte) (int64, []byte) {
	ctx, committer := sqls.Tx(ctx)
	defer committer()
	return models.UserOperator.Create(ctx, name, password)
}
