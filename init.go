package suna

import (
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"log"
)

type InitOption struct {
	Config        *config.Suna
	Authenticator auth.Authenticator
}

var disableReservedKeysWarning bool

func DisableReservedKeysWarning() {
	disableReservedKeysWarning = true
}

func doReservedKeysWarning() {
	if disableReservedKeysWarning {
		return
	}
	log.Printf(
		"suna: reserved fasthttp.RequestCtx.UserValue keys: `%s`, `%s`\n",
		internal.RCtxSessionKey,
		internal.RCtxUserKey,
	)
}

func Init(opt *InitOption) {
	doReservedKeysWarning()

	internal.Dig.Provide(
		func() *config.Suna {
			if opt == nil || opt.Config == nil {
				log.Fatalln("suna: nil init config")
			}
			return opt.Config
		},
	)

	internal.Dig.Provide(
		func() auth.Authenticator {
			if opt == nil {
				return nil
			}
			return opt.Authenticator
		},
	)

	internal.Dig.Invoke()
}
