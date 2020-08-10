package suna

import (
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/cache"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/middleware"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/rbac"
	"github.com/zzztttkkk/suna/redlock"
	"github.com/zzztttkkk/suna/reflectx"
	"github.com/zzztttkkk/suna/secret"
	"github.com/zzztttkkk/suna/session"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/validator"
	"log"
	"reflect"
	"strings"
)

type InitOption struct {
	Config        *config.Suna
	Authenticator auth.Authenticator
}

func Init(opt *InitOption) {
	internal.Dig.Provide(
		func() *config.Suna {
			if opt == nil {
				log.Println("suna: nil config, use the default value")
				return config.GetDefault()
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
func _LoadSubModules() string {
	buf := strings.Builder{}

	buf.WriteString(reflect.ValueOf(cache.NewLru).String())
	buf.WriteString(reflect.ValueOf(middleware.NewAccessLogger).String())
	buf.WriteString(reflect.ValueOf(output.Error).String())
	buf.WriteString(reflect.ValueOf(rbac.Loader).String())
	buf.WriteString(reflect.ValueOf(reflectx.ExportedKeys).String())
	buf.WriteString(reflect.ValueOf(secret.AesDecrypt).String())
	buf.WriteString(reflect.ValueOf(session.New).String())
	buf.WriteString(reflect.ValueOf(sqls.CreateTable).String())
	buf.WriteString(reflect.ValueOf(validator.RegisterFunc).String())
	buf.WriteString(reflect.ValueOf(redlock.New).String())

	return buf.String()
}
