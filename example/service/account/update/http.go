package update

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/secret"
	"github.com/zzztttkkk/suna/utils"
	"github.com/zzztttkkk/suna/validator"
)

type Form struct {
	Name     string `validator:"optional"`
	Alias    string `validator:"optional"`
	Password []byte `validator:"optional;R<password>"`
	Secret   []byte `validator:"optional;L<64>"`
	Bio      string `validator:"optional;L<-255>"`
}

func HttpHandler(ctx *fasthttp.RequestCtx) {
	user := auth.MustGetUser(ctx)

	form := Form{}
	if !validator.Validate(ctx, &form) {
		return
	}

	kvs := utils.AcquireKvs()
	defer kvs.Free()

	if len(form.Alias) > 0 {
		kvs.Append("alias", form.Alias)
	}

	if len(form.Name) > 0 {
		kvs.Append("name", form.Name)
	}

	if len(form.Password) > 0 {
		if len(form.Secret) != 64 {
			output.Error(ctx, output.HttpErrors[fasthttp.StatusForbidden])
			return
		}
		kvs.Append("password", secret.Calc(form.Password))
	}

	if len(form.Bio) > 0 {
		kvs.Append("bio", form.Bio)
	}

	if !Do(ctx, user.GetId(), form.Secret, kvs) {
		output.Error(ctx, output.HttpErrors[fasthttp.StatusBadRequest])
	}
}
