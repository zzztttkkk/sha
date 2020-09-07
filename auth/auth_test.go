package auth

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"testing"
)

func TestGetUser(t *testing.T) {
	authenticatorV = AuthenticatorFunc(
		func(ctx *fasthttp.RequestCtx) (User, bool) {
			return nil, false
		},
	)

	ctx := &fasthttp.RequestCtx{}
	fmt.Println(GetUser(ctx))
	fmt.Println(GetUser(ctx))
}
