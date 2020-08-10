package rbac

import (
	"context"
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
	Del = modifyType(-1)
	Add = modifyType(1)
)

func (v modifyType) String() string {
	switch v {
	case Add:
		return "add"
	case Del:
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

			l := cfg.SqlLeader()
			if l == nil {
				log.Println("suna.rbac: init error")
				return
			}
		},
	)
}
