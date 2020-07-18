package login

import (
	"context"

	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/snow/examples/blog/backend/models"
	"github.com/zzztttkkk/snow/output"
)

func DoLogin(ctx context.Context, name, password []byte, days int) (string, error) {
	uid, ok := models.UserOperator.AuthByName(ctx, name, password)
	if !ok {
		return "", output.StdErrors[fasthttp.StatusUnauthorized]
	}
	return models.UserOperator.Dump(uid, days), nil
}
