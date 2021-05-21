package utils

import (
	"bytes"
	"encoding/hex"
	"hash"
	"sync"
)

type _HashWrapper struct {
	hash.Hash
	sumDist []byte
	hexDist []byte
}

type HashPool struct {
	secret      []byte
	constructor func() hash.Hash
	pool        sync.Pool
	size        int
}

func NewHashPoll(constructor func() hash.Hash, secret []byte) *HashPool {
	ret := &HashPool{
		secret:      secret,
		constructor: constructor,
	}
	ret.pool.New = func() interface{} { return nil }
	ret.size = constructor().Size()
	return ret
}

func (hp *HashPool) get() *_HashWrapper {
	vi := hp.pool.Get()
	if vi != nil {
		return vi.(*_HashWrapper)
	}
	return &_HashWrapper{
		Hash:    hp.constructor(),
		sumDist: make([]byte, 0, hp.size),
		hexDist: make([]byte, hp.size*2),
	}
}

func (hp *HashPool) put(v *_HashWrapper) {
	v.Reset()
	v.sumDist = v.sumDist[:0]
	hp.pool.Put(v)
}

func (hp *HashPool) Sum(v []byte, dist []byte) []byte {
	h := hp.get()
	defer hp.put(h)

	_, _ = h.Write(hp.secret)
	_, _ = h.Write(v)

	hex.Encode(dist, h.Sum(h.sumDist))
	return dist
}

func (hp *HashPool) Equal(v []byte, h []byte) bool {
	if len(h) != hp.size*2 {
		return false
	}

	hw := hp.get()
	defer hp.put(hw)
	_, _ = hw.Write(hp.secret)
	_, _ = hw.Write(v)
	hex.Encode(hw.hexDist, hw.Sum(hw.sumDist))
	return bytes.Equal(hw.hexDist, h)
}
