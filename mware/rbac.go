package mware

import (
	"github.com/valyala/fasthttp"
	rbacp "github.com/zzztttkkk/snow/rbac"
)

func NewPermissionChecker(rbac rbacp.Rbac, permissions ...string) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

	}
}
