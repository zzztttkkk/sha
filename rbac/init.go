package rbac

import "github.com/zzztttkkk/suna/ini"

func Init(conf *ini.Config) {
	MaxPrivateSubjectId = conf.GetIntOr("rbac.max_private_subject", 0)
}
