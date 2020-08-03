package secret

import (
	"fmt"
	"github.com/zzztttkkk/suna/config"
)

var gSecretKey []byte

func Init(conf *config.Type) {
	gSecretKey = []byte(conf.Secret.Key)

	hashMethod := conf.Secret.HashAlgorithm
	if len(hashMethod) < 1 {
		hashMethod = "sha256-512"
	}
	Default = hashMap[hashMethod]
	if Default == nil {
		panic(fmt.Errorf("suna.secret: unknown hash method `%s`", hashMethod))
	}
}
