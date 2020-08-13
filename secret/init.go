package secret

import (
	"fmt"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

var gSecretKey []byte

func init() {
	internal.Dig.LazyInvoke(
		func(conf *config.Suna) {
			gSecretKey = []byte(conf.Secret.Key)
			hashMethod := conf.Secret.HashAlgorithm
			if len(hashMethod) < 1 {
				hashMethod = "sha256-512"
			}
			_Default = hashMap[hashMethod]
			if _Default == nil {
				panic(fmt.Errorf("suna.secret: unknown hash method `%s`", hashMethod))
			}
		},
	)
}
