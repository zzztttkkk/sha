package secret

import (
	"fmt"

	"github.com/zzztttkkk/suna/ini"
)

var gSecretKey []byte

func Init(conf *ini.Ini) {
	gSecretKey = []byte(conf.GetOr("secret.key", ""))

	hashMethod := conf.GetOr("secret.hash", "sha256-512")
	Default = hashMap[hashMethod]
	if Default == nil {
		panic(fmt.Errorf("suna.secret: unknown hash method `%s`", hashMethod))
	}
}
