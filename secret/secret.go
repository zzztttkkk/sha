package secret

import (
	"fmt"

	"github.com/zzztttkkk/snow/ini"
)

var gSecretKey []byte

func Init(conf *ini.Config) {
	gSecretKey = conf.GetMust("secret.key")

	hashMethod := conf.GetOr("secret.hash", "sha256-512")
	Default = hashMap[hashMethod]
	if Default == nil {
		panic(fmt.Errorf("snow.secret: unknown hash method `%s`", hashMethod))
	}
}
