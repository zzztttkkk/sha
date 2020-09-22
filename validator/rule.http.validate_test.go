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
		FIDs      []int64 `validator:"fid:S<3>"`
		Float     float64
	}
	fmt.Print(GetRules(Form{}).NewDoc("").Document())

	ctx := fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://localhost:8080/?password=123456&name=ztk&fid=1&fid=22&fid=3&float=0.123")

	form := Form{}
	if !Validate(&ctx, &form) {
		t.Fatalf("%s", string(ctx.Response.Body()))
		return
	}
	fmt.Println(string(form.Password), string(form.Name), form.FIDs)
	fmt.Println(form)
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
	fmt.Println(form.A.ToCollection().MustGet("a.b.c.0"))

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

func TestN(t *testing.T) {
	type A struct {
		Id int64
	}

	type B struct {
		Name string
	}

	type C struct {
		A
		b    B
		Info string
	}

	fmt.Println(GetRules(C{}).NewDoc("").Document())
}
