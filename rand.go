package suna

import (
	"hash/fnv"
	"math/rand"
	"regexp"
	"time"
)

var randSecretKey []byte
var defaultRandBytesPool []byte
var source rand.Source

func init() {
	emptyRegexp := regexp.MustCompile(`\W+`)

	for i := 1; i < 256; i++ {
		defaultRandBytesPool = append(defaultRandBytesPool, uint8(i))
	}

	defaultRandBytesPool = emptyRegexp.ReplaceAll(defaultRandBytesPool, []byte(""))

	dig.Append(
		func(conf *Config) {
			randSecretKey = conf.SecretKey
			h := fnv.New32a()
			_, _ = h.Write(randSecretKey)
			source = rand.NewSource(time.Now().UnixNano() * int64(h.Sum32()))
		},
	)
}

func RandBytes(size int, pool []byte) []byte {
	if len(pool) < 1 {
		pool = defaultRandBytesPool
	}
	var rv []byte
	ps := int64(len(pool))
	for i := 0; i < size; i++ {
		rv = append(rv, pool[source.Int63()%ps])
	}
	return rv
}
