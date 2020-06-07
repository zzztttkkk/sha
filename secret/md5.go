package secret

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
	"sync"
)

var md5Pool = sync.Pool{New: func() interface{} { return md5.New() }}

func Md5Calc(txt []byte) []byte {
	h := md5Pool.Get().(hash.Hash)
	v := make([]byte, 32)

	_, _ = h.Write(txt)

	defer func() {
		h.Reset()
		md5Pool.Put(h)
	}()

	hex.Encode(v, h.Sum(nil))
	return v
}

func Md5CalcWithSecret(txt []byte) []byte {
	return Md5Calc(append(txt, secretKey...))
}
