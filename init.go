package suna

import (
	"fmt"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/utils"
	"log"
	"strings"
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

	q := func(sl []string) []string {
		return utils.SMap(
			sl,
			func(_ int, s string) string {
				return fmt.Sprintf(`"%s"`, s)
			},
		)
	}

	log.Printf(
		"suna: reserved `fasthttp.RequestCtx.UserValue` keys: %s\n",
		strings.Join(
			q(
				[]string{
					internal.RCtxSessionKey,
					internal.RCtxUserKey,
					internal.RCtxRouterPathParam,
				},
			),
			", ",
		),
	)
	log.Printf(
		"suna: reserved `suna.session.Session` keys: %s\n",
		strings.Join(
			q(
				[]string{
					internal.SessionExistsKey,
				},
			),
			", ",
		),
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
