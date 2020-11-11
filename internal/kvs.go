package internal

import (
	"bytes"
	"strings"
	"sync"
)

type _KvItem struct {
	invalid bool
	key     []byte
	val     []byte
}

func (item *_KvItem) makeInvalid() {
	item.invalid = true
	item.key = item.key[:0]
	item.val = item.val[:0]
}

type Kvs struct {
	lst          []_KvItem
	invalidItems []int
}

func (kvs *Kvs) String() string {
	buf := strings.Builder{}

	kvs.EachItem(
		func(k, v []byte) bool {
			buf.WriteByte('`')
			buf.Write(k)
			buf.WriteByte('`')
			buf.WriteByte(':')
			buf.WriteByte('`')
			buf.Write(v)
			buf.WriteByte('`')
			buf.WriteByte(';')
			buf.WriteByte('\n')
			return true
		},
	)
	return buf.String()
}

var kvsPool = sync.Pool{New: func() interface{} { return &Kvs{} }}

func AcquireKvs() *Kvs {
	return kvsPool.Get().(*Kvs)
}

func ReleaseKvs(kvs *Kvs) {
	kvs.Reset()
	kvsPool.Put(kvs)
}

func (kvs *Kvs) Append(k, v []byte) {
	invItemSize := len(kvs.invalidItems)
	var item *_KvItem
	if invItemSize > 0 {
		ind := kvs.invalidItems[invItemSize-1]
		kvs.invalidItems = kvs.invalidItems[:invItemSize-1]

		item = &(kvs.lst[ind])
		item.invalid = false
	} else {
		kvs.lst = append(kvs.lst, _KvItem{})
		item = &(kvs.lst[len(kvs.lst)-1])
	}
	item.key = append(item.key, k...)
	item.val = append(item.val, v...)
}

func (kvs *Kvs) Set(k, v []byte) {
	kvs.Del(k)
	kvs.Append(k, v)
}

func (kvs *Kvs) Del(k []byte) {
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if bytes.Equal(item.key, k) {
			item.makeInvalid()
			kvs.invalidItems = append(kvs.invalidItems, i)
		}
	}
}

func (kvs *Kvs) Get(k []byte) ([]byte, bool) {
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if bytes.Equal(item.key, k) {
			return item.val, true
		}
	}
	return nil, false
}

func (kvs *Kvs) GetAll(k []byte) [][]byte {
	var rv [][]byte

	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if bytes.Equal(item.key, k) {
			rv = append(rv, item.val[:])
		}
	}
	return rv
}

func (kvs *Kvs) GetAllRef(k []byte) []*[]byte {
	var rv []*[]byte

	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if bytes.Equal(item.key, k) {
			rv = append(rv, &item.val)
		}
	}
	return rv
}

func (kvs *Kvs) EachKey(visitor func(k []byte) bool) {
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if !visitor(item.key) {
			break
		}
	}
}

func (kvs *Kvs) EachValue(visitor func(v []byte) bool) {
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if !visitor(item.val) {
			break
		}
	}
}

func (kvs *Kvs) EachItem(visitor func(k, v []byte) bool) {
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if !visitor(item.key, item.val) {
			break
		}
	}
}

func (kvs *Kvs) Reset() {
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		item.makeInvalid()
		kvs.invalidItems = append(kvs.invalidItems, i)
	}
}
