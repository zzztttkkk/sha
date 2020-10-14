package rbac

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/router"
	"github.com/zzztttkkk/suna/validator"
)

// path: /log/list
func init() {
	type Form struct {
		Begin     int64    `validate:"optional"`
		End       int64    `validate:"optional"`
		Names     []string `validate:"optional"`
		Operators []int64  `validate:"optional"`

		Cursor int64 `validate:"optional"`
		Limit  int64 `validate:"V<10-30>;D<10>;optional"`
		Asc    bool  `validate:"D<true>;optional"`
	}

	loader.Http(
		func(router router.Router) {
			router.GETWithDoc(
				"/log/list",
				newPAllPermChecker(
					"rbac.log.read",
					func(ctx *fasthttp.RequestCtx) {
						form := Form{}
						if !validator.Validate(ctx, &form) {
							return
						}

						output.MsgOK(
							ctx,
							LogOperator.List(
								ctx,
								form.Begin, form.End,
								form.Names, form.Operators,
								form.Asc,
								form.Cursor, form.Limit,
							),
						)
					},
				),
				validator.MakeDoc(Form{}, "list rbac operation logging"),
			)
		},
	)
}
