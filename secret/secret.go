package secret

import (
	"fmt"
	"github.com/zzztttkkk/snow/ini"
)

var secretKey []byte

func Init() {
	secretKey = ini.GetMust("app.secret")

	hashMethod := ini.GetOr("app.hash", "sha256-512")
	Hash = hashMap[hashMethod]
	if Hash == nil {
		panic(fmt.Errorf("snow.secret: unknown hash method `%s`", hashMethod))
	}
}
