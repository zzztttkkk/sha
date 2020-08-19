package service

import (
	"github.com/zzztttkkk/suna.example/service/account"
	"github.com/zzztttkkk/suna/utils"
)

var Loader = utils.NewLoader()

func init() {
	Loader.AddChild("/account", account.Loader)
}
