package services

import (
	"github.com/zzztttkkk/snow"
	"github.com/zzztttkkk/snow/examples/blog/backend/services/account"
)

var Loader = snow.NewLoader()

func init() {
	Loader.AddChild("account", account.Loader)
}
