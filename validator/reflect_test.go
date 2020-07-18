package validator

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestValidate(t *testing.T) {
	type Form struct {
		Password  []byte `validator:":R<password>"`
		Name      []byte `validator:":F<username>"`
		KeepLogin bool   `validator:"kl:optional"`
		ExtInfo   map[string]interface{}
		FIDs      []int64 `validator:"fid:"`
	}
	fmt.Println(getRules(reflect.TypeOf(Form{})))

	ctx := fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://localhost:8080/?extinfo=%7B%22a%22%3A23%2C%20%22b%22%3A%2045%7D&password=123456&name=ztk&fid=1&fid=2&fid=3")

	form := Form{}
	if !Validate(&ctx, &form) {
		t.Fail()
		return
	}
	fmt.Println(string(form.Password), string(form.Name), form.ExtInfo, form.FIDs)
}
