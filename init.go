package suna

import (
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/cache"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/jsonx"
	"github.com/zzztttkkk/suna/middleware"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/rbac"
	"github.com/zzztttkkk/suna/redlock"
	"github.com/zzztttkkk/suna/reflectx"
	"github.com/zzztttkkk/suna/secret"
	"github.com/zzztttkkk/suna/session"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/sqls/builder"
	"github.com/zzztttkkk/suna/validator"
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
	log.Printf(
		"suna: reserved suna.session.Session keys: `%s`, `%s`, `%s`\n",
		internal.SessionExistsKey,
		internal.SessionCaptchaIdKey,
		internal.SessionCaptchaUnixKey,
	)
}

func Init(opt *InitOption) {
	doReservedKeysWarning()

	internal.Dig.Provide(
		func() *config.Suna {
			if opt == nil {
				log.Fatalln("suna: nil init option")
			}
			if opt.Config == nil {
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

	_LoadSubModules()
	internal.Dig.Invoke()
}

// trigger internal.LazyInvoke
func _LoadSubModules() {
	internal.Dig.Index(cache.NewLru)
	internal.Dig.Index(middleware.NewAccessLogger)
	internal.Dig.Index(output.Error)
	internal.Dig.Index(rbac.Loader)
	internal.Dig.Index(reflectx.ExportedKeys)
	internal.Dig.Index(secret.AesDecrypt)
	internal.Dig.Index(session.New)
	internal.Dig.Index(sqls.CreateTable)
	internal.Dig.Index(builder.AndConditions)
	internal.Dig.Index(validator.RegisterFunc)
	internal.Dig.Index(redlock.New)
	internal.Dig.Index(jsonx.Marshal)
}
