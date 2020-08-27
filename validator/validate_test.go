package validator

import (
	"fmt"
	"github.com/zzztttkkk/suna/jsonx"
	"regexp"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestValidate(t *testing.T) {
	var spaceReg = regexp.MustCompile(`\s+`)
	var pwdReg = regexp.MustCompile(`\w{6,}`)

	RegisterFunc(
		"username", func(data []byte) ([]byte, bool) {
			return spaceReg.ReplaceAll(data, nil), true
		},
		"remove all space",
	)

	RegisterRegexp("password", pwdReg)

	type Form struct {
		Password  []byte  `validator:":R<password>"`
		Name      []byte  `validator:":F<username>"`
		KeepLogin bool    `validator:"kl:optional"`
		FIDs      []int64 `validator:"fid:S<7>"`
	}
	fmt.Println(GetRules(Form{}).NewDoc(""))

	ctx := fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://localhost:8080/?password=123456&name=ztk&fid=1&fid=2&fid=3")

	form := Form{}
	if !Validate(&ctx, &form) {
		t.Fatalf("%s", string(ctx.Response.Body()))
		return
	}
	fmt.Println(string(form.Password), string(form.Name), form.FIDs)
}

func TestJsonRequest(t *testing.T) {
	type Form struct {
		A jsonx.Object
	}

	ctx := fasthttp.RequestCtx{}
	ctx.Request.SetBody([]byte(`{"a":{"b":{"c":[1, 2, 3]}}}`))
	ctx.Request.Header.SetContentType("application/json")

	form := Form{}
	if !Validate(&ctx, &form) {
		t.Fatalf("%s", string(ctx.Response.Body()))
		return
	}
	fmt.Println(form.A.Get("a.b.c.0"))

	type Info struct {
		A struct {
			B struct {
				C []int64
			}
		}
	}

	info := Info{}
	if !ValidateJson(&ctx, &info) {
		t.Fatalf("%s", string(ctx.Response.Body()))
		return
	}
	fmt.Println(info)
}
