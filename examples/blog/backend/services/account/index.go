package account

import (
	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend/services/account/login"
	"github.com/zzztttkkk/snow/router"
)

var Loader = snow.NewLoader()

func init() {
	Loader.Http(
		func(router *router.Router) {
			router.POST("/register", Register)
			router.GET("/login", login.Handler)
			router.POST("/unregister", Unregister)
		},
	)
}
