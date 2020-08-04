package secret

import (
	"bytes"
	"fmt"
	"github.com/zzztttkkk/suna/config"
	"io/ioutil"
	"os"
)

var gSecretKey []byte

func Init(conf *config.Type) {
	gSecretKey = []byte(conf.Secret.Key)
	if bytes.HasPrefix(gSecretKey, []byte("file://")) {
		f, e := os.Open(string(gSecretKey[7:]))
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
	Default = hashMap[hashMethod]
	if Default == nil {
		panic(fmt.Errorf("suna.secret: unknown hash method `%s`", hashMethod))
	}
}
