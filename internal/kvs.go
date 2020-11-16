package internal

import (
	"bytes"
	"strings"
	"sync"
)

type KvItem struct {
	invalid bool
	Key     []byte
	Val     []byte
}

func (item *KvItem) makeInvalid() {
	item.invalid = true
	item.Key = item.Key[:0]
	item.Val = item.Val[:0]
}

type Kvs struct {
	lst          []KvItem
	invalidItems []int
	size         int
}

func (kvs *Kvs) Size() int {
	return kvs.size
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

func (kvs *Kvs) Append(k, v []byte) *KvItem {
	invItemSize := len(kvs.invalidItems)
	var item *KvItem
	if invItemSize > 0 {
		ind := kvs.invalidItems[invItemSize-1]
		kvs.invalidItems = kvs.invalidItems[:invItemSize-1]

		item = &(kvs.lst[ind])
		item.invalid = false
	} else {
		kvs.lst = append(kvs.lst, KvItem{})
		item = &(kvs.lst[len(kvs.lst)-1])
	}
	item.Key = append(item.Key, k...)
	item.Val = append(item.Val, v...)
	kvs.size++
	return item
}

func (kvs *Kvs) Set(k, v []byte) *KvItem {
	kvs.Del(k)
	return kvs.Append(k, v)
}

func (kvs *Kvs) Del(k []byte) {
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if bytes.Equal(item.Key, k) {
			item.makeInvalid()
			kvs.invalidItems = append(kvs.invalidItems, i)
			kvs.size--
		}
	}
}

func (kvs *Kvs) Get(k []byte) ([]byte, bool) {
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if bytes.Equal(item.Key, k) {
			return item.Val, true
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
		if bytes.Equal(item.Key, k) {
			rv = append(rv, item.Val)
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
		if bytes.Equal(item.Key, k) {
			rv = append(rv, &item.Val)
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
		if !visitor(item.Key) {
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
		if !visitor(item.Val) {
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
		if !visitor(item.Key, item.Val) {
			break
		}
	}
}

func (kvs *Kvs) Reset() {
	if len(kvs.invalidItems) == len(kvs.lst) {
		return
	}
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		item.makeInvalid()
		kvs.invalidItems = append(kvs.invalidItems, i)
		kvs.size = 0
	}
}
