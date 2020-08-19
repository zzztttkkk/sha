package login

import (
	"context"
	"github.com/zzztttkkk/suna.example/model"

	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/suna/output"
)

func Do(ctx context.Context, name, password []byte, seconds int64) (string, error) {
	uid, ok := model.UserOperator.AuthByName(ctx, name, password)
	if !ok {
		return "", output.HttpErrors[fasthttp.StatusUnauthorized]
	}
	return model.UserOperator.Dump(uid, seconds), nil
}
