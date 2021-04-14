package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"github.com/zzztttkkk/sha/validator"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Time time.Time

func (t *Time) FormValue(data []byte) bool {
	v, e := strconv.ParseInt(utils.S(data), 10, 64)
	if e != nil {
		return false
	}
	if v < 6000 {
		return false
	}
	*t = Time(time.Unix(v-6000, 0))
	return true
}

func (t *Time) String() string {
	return (*time.Time)(t).String()
}

type TV01Form struct {
	//NumbersPtr *[]int64
	//Numbers    []int64 `validator:",w=body"`
	//StrPtr     *string
	//Str        string
	CFTime       Time
	CFTimePtr    *Time
	CFTimePtrPtr **Time
}

func (t *TV01Form) String() string {
	sb := strings.Builder{}

	//sb.WriteString("NumbersPtr: ")
	//if t.NumbersPtr == nil {
	//	sb.WriteString("nil\n")
	//} else {
	//	sb.WriteString(fmt.Sprintf("%v\n", *t.NumbersPtr))
	//}
	//
	//sb.WriteString(fmt.Sprintf("Numbers: %v\n", t.Numbers))
	//
	//sb.WriteString("StrPtr: ")
	//if t.StrPtr == nil {
	//	sb.WriteString("nil\n")
	//} else {
	//	sb.WriteString(fmt.Sprintf("%p<%s>\n", t.StrPtr, *t.StrPtr))
	//}
	//
	//sb.WriteString(fmt.Sprintf("Str: %s\n", t.Str))

	sb.WriteString(fmt.Sprintf("CFTime: %s\n", &t.CFTime))

	sb.WriteString("CFTimePtr: ")
	if t.CFTimePtr == nil {
		sb.WriteString("nil\n")
	} else {
		sb.WriteString(fmt.Sprintf("%s\n", t.CFTimePtr))
	}

	return sb.String()
}

func (t *TV01Form) DefaultNumbers() interface{} { return []int64{1, 2, 3} }

func TestValidator(t *testing.T) {
	ListenAndServe(
		"127.0.0.1:5986",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			var form TV01Form

			if err := ctx.ValidateForm(&form); err != nil {
				ctx.Response.SetStatusCode(err.StatusCode())
				_, _ = ctx.WriteString(err.Error())
				return
			}

			fmt.Print(&form)
		}),
	)
}

func TestRequestCtx_ValidateJSON(t *testing.T) {
	validator.RegisterRegexp("name", regexp.MustCompile("\\w{5}"))

	type Form struct {
		Num  int64  `json:"num" validator:",v=10-50"`
		Name string `json:"name" validator:",r=name"`
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
		Password validator.Password `validator:"pwd"`
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
