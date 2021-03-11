package sha

import (
	"fmt"
	"testing"
)

type TV01Form struct {
	Numbers []int64 `validator:",w=body"`
}

func (t *TV01Form) DefaultNumbers() interface{} { return []int64{1, 2, 3} }

func TestValidator(t *testing.T) {
	ListenAndServe(
		"127.0.0.1:5986",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			var form TV01Form
			ctx.MustValidate(&form)
			form.Numbers[0] = 45
			fmt.Printf("%v %p\n", form.Numbers, &form.Numbers[0])
		}),
	)
}
