package secret

import (
	crand "crypto/rand"
	"math"
	"math/big"
	"math/rand"
)

var AsciiLowerLetters = []byte("abcdefghijklmnopqrstuvwxyz")
var AsciiUpperLetters = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
var AsciiLetters = make([]byte, 0)
var Digits = []byte("0123456789")
var Asciis = make([]byte, 0)
var HexDigits = make([]byte, 0)
var Base64 = make([]byte, 0)
var Base58 = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func init() {
	AsciiLetters = append(AsciiLetters, append(AsciiLowerLetters, AsciiUpperLetters...)...)
	Asciis = append(Asciis, append(AsciiLetters, Digits...)...)
	HexDigits = append(HexDigits, append(Digits, []byte("abcdefABCDEF")...)...)
	Base64 = append(Base64, append(AsciiLetters, Digits...)...)
}

var SeedPoolSize uint32 = 512
var _max = big.Int{}
var seedChan chan int64

func genSeed() {
	rv, err := crand.Int(crand.Reader, &_max)
	if err != nil {
		panic(err)
	}
	seedChan <- rv.Int64()
}

func init() {
	_max.SetInt64(math.MaxInt64)
	seedChan = make(chan int64, SeedPoolSize)

	var i uint32 = 0
	for ; i < SeedPoolSize/5; i++ {
		genSeed()
	}

	go func() {
		for {
			genSeed()
		}
	}()
}

func RandBytes(n int, pool []byte) []byte {
	if pool == nil {
		pool = Asciis
	}

	seed := <-seedChan

	rand.Seed(seed)

	c := make([]byte, n)
	rand.Read(c)

	l := len(pool)
	for i, b := range c {
		c[i] = pool[int(b)%l]
	}
	return c
}
