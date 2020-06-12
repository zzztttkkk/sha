package account

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/examples/blog/backend/models"
	"github.com/zzztttkkk/snow/mware"
	"github.com/zzztttkkk/snow/mware/ctxs"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/validator"
)

type LoginForm struct {
	RegisterForm
	Days int `validator:":optional;D<0>;V<0-30>"`
}

func Login(ctx *fasthttp.RequestCtx) {
	session := ctxs.Session(ctx)
	if !session.CaptchaVerify(ctx) {
		output.StdError(ctx, fasthttp.StatusBadRequest)
		return
	}

	form := LoginForm{}
	if !validator.Validate(ctx, &form) {
		return
	}
	token := doLogin(ctx, form.Name, form.Password, form.Days)
	output.MsgOk(ctx, token)
}

func doLogin(ctx context.Context, name, password []byte, days int) string {
	uid, ok := models.UserOperator.AuthByName(ctx, name, password)
	if !ok {
		panic(output.StdErrors[fasthttp.StatusUnauthorized])
	}
	return mware.AuthToken(uid, days*86400, nil)
}
