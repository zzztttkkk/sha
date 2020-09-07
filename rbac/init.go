package rbac

import (
	"context"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/utils"
	"log"
)

var tablePrefix = "rbac_"
var lazier = utils.NewLazyExecutor()
var initPriority = utils.NewPriority(1)
var permTablePriority = initPriority.Incr()
var rbacPriority = permTablePriority.Incr().Incr()
var loader = utils.NewLoader()

type modifyType int

const (
	_Del = modifyType(-1)
	_Add = modifyType(1)
)

func (v modifyType) String() string {
	switch v {
	case _Add:
		return "add"
	case _Del:
		return "del"
	default:
		return ""
	}
}

func Loader() *utils.Loader {
	lazier.RegisterWithPriority(
		func(kwargs utils.Kwargs) { Load(context.Background()) },
		rbacPriority,
	)
	lazier.Execute(nil)
	return loader
}

var cfg *config.Suna

func init() {
	internal.Dig.LazyInvoke(
		func(confV *config.Suna) {
			cfg = confV
			tablePrefix = confV.Rbac.TablenamePrefix

			if cfg.GetSqlLeader() == nil {
				log.Fatalln("suna.rbac: nil sql")
				return
			}
			if !auth.IsAvailable() {
				log.Fatalln("suna.rbac: auth is unavailable")
				return
			}
		},
	)
}
