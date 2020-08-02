package conf

import (
	"github.com/zzztttkkk/suna/auth"
	"time"
)

type Conf struct {
	Authenticator auth.Authenticator

	Secret struct {
		Key           string
		HashAlgorithm string
	}

	Session struct {
		Header string
		Cookie string
		Prefix string
	}

	Errors struct {
		MaxDepth int
	}

	Sql struct {
		Driver          string
		Leader          string
		Followers       []string
		EnumCacheMaxAge time.Duration
	}

	Redis struct {
		Mode  string
		Nodes []string
	}
}

func FromFiles() {

}
