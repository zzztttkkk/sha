package utils

import (
	"crypto/rand"
	"math"
	"math/big"
	mrandlib "math/rand"
)

func MathRandSeed() {
	max := big.NewInt(math.MaxInt64)
	v, e := rand.Int(rand.Reader, max)
	if e != nil {
		panic(e)
	}
	mrandlib.Seed(v.Int64())
}

func init() {
	MathRandSeed()
}

var defaultRandBytesPool = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var Base58BytesPool = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func RandByte(pool []byte) byte {
	if pool == nil {
		pool = defaultRandBytesPool
	}
	return pool[mrandlib.Intn(len(pool))]
}

func RandBytes(dist, pool []byte) {
	for i := 0; i < len(dist); i++ {
		dist[i] = RandByte(pool)
	}
}
