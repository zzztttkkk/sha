package sha

import (
	"errors"
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"github.com/zzztttkkk/sha/validator"
	"regexp"
	"strconv"
	"testing"
	"time"
)

type Time time.Time

func (t *Time) FromBytes(data []byte) error {
	v, e := strconv.ParseInt(utils.S(data), 10, 64)
	if e != nil {
		return e
	}
	*t = Time(time.Unix(v, 0))
	return nil
}

func (t *Time) Validate() error {
	if (*time.Time)(t).Before(time.Date(1900, 12, 12, 12, 12, 12, 12, time.UTC)) {
		return errors.New("unvalidated time")
	}
	return nil
}

func (t *Time) String() string {
	return (*time.Time)(t).String()
}

type TV01Form struct {
	NumbersPtr *[]int64 `vld:"a,optional"`
	Numbers    []int64  `vld:"b"`
	StrPtr     *string  `vld:"c,optional"`
	Str        string   `vld:"d,optional"`
	Time       Time     `vld:"e,optional"`
	TimePtr    *Time    `vld:"f,optional"`
	TimePtrs   []*Time  `vld:"g,optional"`
	Times      []Time   `vld:"h,optional"`
}

func (t *TV01Form) Default(name string) func() interface{} {
	switch name {
	case "Numbers":
		return func() interface{} { return []int64{1, 2, 3} }
	default:
		return nil
	}
}

func TestValidator(t *testing.T) {
	ListenAndServe(
		"127.0.0.1:5986",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			var form TV01Form

			if err := ctx.ValidateForm(&form); err != nil {
				ctx.Response.SetStatusCode(err.StatusCode())
				_ = ctx.WriteString(err.Error())
				return
			}

			fmt.Printf("%+v\r\n%p\r\n", &form, form.Numbers)
		}),
	)
}

func TestRequestCtx_ValidateJSON(t *testing.T) {
	validator.RegisterRegexp("name", regexp.MustCompile("\\w{5}"))

	type Form struct {
		Num  int64  `json:"num" vld:",v=10-50"`
		Name string `json:"name" vld:",r=name"`
	}

	rctx := &RequestCtx{}
	rctx.Request.Header().SetContentType(MIMEJson)
	_, _ = rctx.Request._HTTPPocket.Write([]byte(`{"num":45, "name":"MOON1"}`))

	var form Form
	rctx.MustValidateJSON(&form)

	fmt.Printf("%+v\n", &form)
}

func TestPwd(t *testing.T) {
	type Form struct {
		Password validator.Password `vld:"pwd"`
	}

	mux := NewMux(nil)
	mux.HTTP(MethodGet, "/", RequestHandlerFunc(func(ctx *RequestCtx) {
		var form Form
		if err := ctx.ValidateForm(&form); err != nil {
			ctx.SetError(err)
			return
		}

		h, _ := form.Password.BcryptHash(-1)
		fmt.Println(form.Password.MatchTo(h))
	}))

	ListenAndServe("", mux)
}
