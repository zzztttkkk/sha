package builder

import (
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

var isPostgres bool

func init() {
	internal.Dig.LazyInvoke(
		func(conf *config.Suna) {
			if conf == nil || conf.GetSqlLeader() == nil {
				return
			}
			isPostgres = conf.GetSqlLeader().DriverName() == "postgres"
		},
	)
}
