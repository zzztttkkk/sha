package secret

import (
	"crypto/rand"
	"math"
	"math/big"
	mrand "math/rand"
	"time"
)

var src mrand.Source

func init() {
	nBig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(err)
	}
	src = mrand.NewSource(nBig.Int64())
	src.Seed(time.Now().UnixNano())
}

var defaultPool = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandBytes(size int, pool []byte) []byte {
	if pool == nil {
		pool = defaultPool
	}
	rv := make([]byte, size, size)
	_l := int64(len(pool))
	for i := 0; i < size; i++ {
		rv[i] = pool[src.Int63()%_l]
	}
	return rv
}
