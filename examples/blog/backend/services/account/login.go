package account

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/examples/blog/backend/models"
	"github.com/zzztttkkk/snow/mware"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/secret"
	"github.com/zzztttkkk/snow/validator"
	"time"
)

type LoginForm struct {
	RegisterForm
	Days int `validator:":optional;D<0>;V<0-30>"`
}

func Login(ctx *fasthttp.RequestCtx) {
	session := mware.GetSessionMust(ctx)
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
	if ok {
		panic(output.StdErrors[fasthttp.StatusUnauthorized])
	}
	return secret.JwtEncode(
		jwt.MapClaims{
			"exp":  time.Now().Unix() + int64(86400)*int64(days),
			"name": name,
			"uid":  uid,
		},
	)
}
