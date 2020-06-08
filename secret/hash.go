package secret

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"github.com/zzztttkkk/snow/utils"
	"hash"
	"sync"
)

type _HashT struct {
	pool   *sync.Pool
	bytes  *utils.BytesPool
	size   int
	secret bool
}

func NewHash(fn func() hash.Hash, withSecret bool) *_HashT {
	pool := &sync.Pool{New: func() interface{} { return fn() }}
	t := pool.Get().(hash.Hash)
	size := t.Size() * 2
	return &_HashT{
		pool:   pool,
		bytes:  utils.NewBytesPool(size, size),
		size:   size,
		secret: withSecret,
	}
}

func (h *_HashT) Calc(v []byte) []byte {
	method := h.pool.Get().(hash.Hash)
	result := make([]byte, h.size)
	_, _ = method.Write(v)
	if h.secret {
		method.Write(secretKey)
	}

	defer func() {
		method.Reset()
		h.pool.Put(method)
	}()
	hex.Encode(result, method.Sum(nil))
	return result
}

func (h *_HashT) Equal(raw []byte, hexBytes []byte) bool {
	method := h.pool.Get().(hash.Hash)
	result := h.bytes.Get()

	method.Write(raw)
	if h.secret {
		method.Write(secretKey)
	}

	defer func() {
		method.Reset()
		h.pool.Put(method)
		h.bytes.Put(result)
	}()

	return bytes.Equal(*result, hexBytes)
}

var Md5 = NewHash(md5.New, true)
var Md5Std = NewHash(md5.New, false)

var Sha256 = NewHash(sha256.New, true)
var Sha256Std = NewHash(sha256.New, false)

var Sha512 = NewHash(sha512.New, true)
var Sha512Std = NewHash(sha512.New, false)

//noinspection GoSnakeCaseUsage
var Sha256_512 = NewHash(sha512.New512_256, true)

//noinspection GoSnakeCaseUsage
var Sha256_512Std = NewHash(sha512.New512_256, false)

//noinspection GoSnakeCaseUsage
var Sha384_512 = NewHash(sha512.New384, true)

//noinspection GoSnakeCaseUsage
var Sha384_512Std = NewHash(sha512.New384, false)

var Hash = Sha256_512

var hashMap = map[string]*_HashT{
	"md5":            Md5,
	"md5-std":        Md5Std,
	"sha256":         Sha256,
	"sha256-std":     Sha256Std,
	"sha512":         Sha512,
	"sha512-std":     Sha512Std,
	"sha256-512":     Sha256_512,
	"sha256-512-std": Sha256_512Std,
	"sha384-512":     Sha384_512,
	"sha384-512-std": Sha384_512Std,
}
