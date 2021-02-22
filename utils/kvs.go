package utils

import (
	"strings"
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
		func(item *KvItem) bool {
			buf.Write(item.Key)
			buf.WriteByte(' ')
			buf.Write(item.Val)
			buf.WriteByte(';')
			buf.WriteByte(' ')
			return true
		},
	)
	buf.WriteByte(']')
	return buf.String()
}

func (kvs *Kvs) Append(k string, v []byte) *KvItem { return kvs.AppendBytes(B(k), v) }

func (kvs *Kvs) AppendBytes(k, v []byte) *KvItem {
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

func (kvs *Kvs) Set(k string, v []byte) {
	kvs.Del(k)
	kvs.Append(k, v)
}

func (kvs *Kvs) Del(k string) {
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if string(item.Key) == k {
			item.makeInvalid()
		}
	}
}

func (kvs *Kvs) Get(k string) ([]byte, bool) {
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if string(item.Key) == k {
			return item.Val, true
		}
	}
	return nil, false
}

func (kvs *Kvs) GetAll(k string) [][]byte {
	var rv [][]byte

	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if string(item.Key) == k {
			rv = append(rv, item.Val)
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

func (kvs *Kvs) EachItem(visitor func(item *KvItem) bool) {
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if !visitor(item) {
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
