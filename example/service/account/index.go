package account

import (
	"github.com/zzztttkkk/router"
	"github.com/zzztttkkk/suna.example/service/account/login"
	"github.com/zzztttkkk/suna.example/service/account/register"
	"github.com/zzztttkkk/suna.example/service/account/unregister"
	"github.com/zzztttkkk/suna.example/service/account/update"
	"github.com/zzztttkkk/suna/utils"
	"github.com/zzztttkkk/suna/validator"
)

var Loader = utils.NewLoader()

func init() {
	Loader.Http(
		func(router router.Router) {
			router.POSTWithDoc(
				"/login",
				login.HttpHandler,
				validator.GetRules(login.Form{}).NewDoc(""),
			)
		},
	)
}

func init() {
	Loader.Http(
		func(router router.Router) {
			router.POSTWithDoc(
				"/register",
				register.HttpHandler,
				validator.GetRules(register.Form{}).NewDoc(""),
			)

			router.POSTWithDoc(
				"/unregister",
				unregister.HttpHandler,
				validator.GetRules(unregister.Form{}).NewDoc(""),
			)

			router.POSTWithDoc(
				"/update",
				update.HttpHandler,
				validator.GetRules(update.Form{}).NewDoc(""),
			)
		},
	)
}
