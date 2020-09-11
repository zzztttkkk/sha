package auth

import (
	"fmt"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestGetUser(_ *testing.T) {
	authenticatorV = AuthenticatorFunc(
		func(ctx *fasthttp.RequestCtx) (User, bool) {
			return nil, false
		},
	)

	ctx := &fasthttp.RequestCtx{}
	fmt.Println(GetUser(ctx))
	fmt.Println(GetUser(ctx))
}
