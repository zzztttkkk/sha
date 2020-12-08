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
	lst []KvItem
}

func (kvs *Kvs) Size() int {
	return len(kvs.lst)
}

func (kvs *Kvs) String() string {
	buf := strings.Builder{}
	buf.WriteString("kvs[")
	kvs.EachItem(
		func(k, v []byte) bool {
			buf.Write(k)
			buf.WriteByte(' ')
			buf.Write(v)
			buf.WriteByte(';')
			buf.WriteByte(' ')
			return true
		},
	)
	buf.WriteByte(']')
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
	var item *KvItem

	s := len(kvs.lst)
	if cap(kvs.lst) > s {
		kvs.lst = kvs.lst[:s+1]
	} else {
		kvs.lst = append(kvs.lst, KvItem{})
	}
	item = &(kvs.lst[len(kvs.lst)-1])
	item.invalid = false
	item.Key = append(item.Key, k...)
	item.Val = append(item.Val, v...)

	return item
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
		if bytes.Equal(item.Key, k) {
			item.makeInvalid()
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

func (kvs *Kvs) GetStr(k string) string {
	v, ok := kvs.Get(B(k))
	if !ok {
		return ""
	}
	return S(v)
}

func (kvs *Kvs) AppendStr(k, v string) {
	kvs.Append(B(k), B(v))
}

func (kvs *Kvs) SetStr(k, v string) {
	kvs.Set(B(k), B(v))
}

func (kvs *Kvs) DelStr(k string) {
	kvs.Del(B(k))
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

func (kvs *Kvs) GetAllStr(k string) []string {
	var ret []string
	for _, v := range kvs.GetAll(B(k)) {
		ret = append(ret, S(v))
	}
	return ret
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
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		item.makeInvalid()
	}
	kvs.lst = kvs.lst[:0]
}
