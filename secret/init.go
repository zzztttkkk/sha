package secret

import (
	"fmt"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

var _SecretKey []byte

func init() {
	internal.Dig.Append(
		func(conf *config.Suna) {
			_SecretKey = []byte(conf.Secret.Key)
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
