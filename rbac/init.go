package rbac

import (
	"context"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/router"
	"github.com/zzztttkkk/suna/utils"
	"log"
	"sync"
)

type _DigLogTableInited int
type _DigPermissionTableInited int
type _DigRoleTableInited int
type _DigUserTableInited int

var dig = utils.NewDigContainer()

var tablePrefix = "rbac_"

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

var once sync.Once
var loader *router.Loader

func Loader() *router.Loader {
	once.Do(
		func() {
			loader = router.NewLoader()
			dig.Provide(func() *router.Loader { return loader }, )
			dig.Append(func(_ _DigUserTableInited) { Load(context.Background()) }, )
			dig.Invoke()
		},
	)
	return loader
}

var cfg *config.Suna

func init() {
	internal.Dig.Append(
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
