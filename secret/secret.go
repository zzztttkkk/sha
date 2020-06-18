package secret

import (
	"fmt"
	"github.com/zzztttkkk/snow/ini"
)

var appSecretKey []byte

func Init() {
	appSecretKey = ini.GetMust("app.secret.main")

	hashMethod := ini.GetOr("app.secret.hash", "sha256-512")
	Default = hashMap[hashMethod]
	if Default == nil {
		panic(fmt.Errorf("snow.secret: unknown hash method `%s`", hashMethod))
	}
}
