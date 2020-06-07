package secret

import (
	"fmt"
	"github.com/zzztttkkk/snow/ini"
	"os"
	"strings"
)

var secretKey []byte

func Init() {
	_sk := ini.MustGet("app.secret")
	var err error

	if strings.HasPrefix(_sk, "file://") {
		var f *os.File
		f, err = os.Open(_sk[7:])
		if err != nil {
			panic(err)
		}

		c := make([]byte, 32)
		var l int
		l, err = f.Read(c)
		if err != nil {
			panic(err)
		}
		for i := 0; i < l; i++ {
			secretKey = append(secretKey, c[i])
		}
	} else {
		secretKey = []byte(_sk)
	}

	if len(secretKey) >= 32 {
		secretKey = secretKey[:32]
	} else if len(secretKey) >= 24 {
		secretKey = secretKey[:24]
	} else if len(secretKey) >= 16 {
		secretKey = secretKey[:16]
	} else {
		for len(secretKey) != 16 {
			secretKey = append(secretKey, ' ')
		}
	}

	hashMethod := ini.GetOr("app.hash", "sha256-512")
	Hash = hashMap[hashMethod]
	if Hash == nil {
		panic(fmt.Errorf("snow.secret: unknown hash method `%s`", hashMethod))
	}
}
