package account

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/examples/blog/backend/models"
	"github.com/zzztttkkk/snow/mware"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/sqls"
	"github.com/zzztttkkk/snow/validator"
)

type RegisterForm struct {
	Name     []byte `validator:":F<username>;L<3-20>"`
	Password []byte `validator:":R<password>"`
}

func Register(ctx *fasthttp.RequestCtx) {
	session := mware.GetSessionMust(ctx)
	if session.CaptchaVerify(ctx) {
		output.StdError(ctx, fasthttp.StatusBadRequest)
		return
	}

	form := RegisterForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	doRegister(ctx, form.Name, form.Password)
}

func doRegister(ctx context.Context, name, password []byte) int64 {
	ctx, committer := sqls.Tx(ctx)
	defer committer()
	return models.UserOperator.Create(ctx, name, password)
}
