package secret

import (
	"bytes"
	"fmt"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"io/ioutil"
	"os"
	"path/filepath"
)

var gSecretKey []byte

func init() {
	internal.Dig.LazyInvoke(
		func(conf *config.Suna) {
			gSecretKey = []byte(conf.Secret.Key)
			if bytes.HasPrefix(gSecretKey, []byte("file://")) {
				fp, e := filepath.Abs(string(gSecretKey[7:]))
				if e != nil {
					panic(e)
				}

				f, e := os.Open(fp)
				if e != nil {
					panic(e)
				}
				defer f.Close()
				data, e := ioutil.ReadAll(f)
				if e != nil {
					panic(e)
				}
				gSecretKey = data
			}

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
