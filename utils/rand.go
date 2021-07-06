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
var defaultRandRunePool = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var Base58RunesPool = []rune("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

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

func RandBase58Bytes(dist []byte) {
	for i := 0; i < len(dist); i++ {
		dist[i] = RandByte(Base58BytesPool)
	}
}

func RandRune(pool []rune) rune {
	if pool == nil {
		pool = defaultRandRunePool
	}
	return pool[mrandlib.Intn(len(pool))]
}

func RandRunes(dist, pool []rune) {
	for i := 0; i < len(dist); i++ {
		dist[i] = RandRune(pool)
	}
}

func RandBase58Runes(dist []rune) {
	for i := 0; i < len(dist); i++ {
		dist[i] = RandRune(Base58RunesPool)
	}
}
