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
