package secret

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"github.com/zzztttkkk/snow/internal"
	"hash"
	"sync"
)

var sha256Pool = sync.Pool{New: func() interface{} { return sha256.New() }}

var _64BytesPool = internal.NewBytesPool(64, 64)

func Sha256Equal(txt []byte, hashVal []byte) bool {
	h := sha256Pool.Get().(hash.Hash)
	buf := _64BytesPool.Get()

	_, _ = h.Write(txt)
	_, _ = h.Write(secretKey)

	defer func() {
		h.Reset()
		sha256Pool.Put(h)
		_64BytesPool.Put(buf)
	}()

	hex.Encode(*buf, h.Sum(nil))
	return bytes.Equal(*buf, hashVal)
}

func Sha256CalcWithSecret(txt []byte) []byte {
	return Sha256Calc(append(txt, secretKey...))
}

func Sha256Calc(txt []byte) []byte {
	h := sha256Pool.Get().(hash.Hash)
	v := make([]byte, 64)

	_, _ = h.Write(txt)

	defer func() {
		h.Reset()
		sha256Pool.Put(h)
	}()

	hex.Encode(v, h.Sum(nil))
	return v
}
