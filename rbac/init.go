package rbac

import (
	"context"
	"github.com/zzztttkkk/suna/ini"
	"github.com/zzztttkkk/suna/utils"
)

var tablePrefix string
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
	return loader
}

var conf *ini.Ini

func Init(confV *ini.Ini) {
	conf = confV
	lazier.Execute(nil)
}
