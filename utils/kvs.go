package utils

import (
	"net/http"
	"net/url"
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

type MultiValueMap map[string][]string

type Kvs struct {
	lst  []KvItem
	size int
}

type _KvsIterable interface {
	EachItem(visitor func(item *KvItem) bool)
}

func (kvs *Kvs) Size() int { return kvs.size }

func (kvs *Kvs) Cap() int { return cap(kvs.lst) }

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

	kvs.size++
	return item
}

func (kvs *Kvs) AppendString(k, v string) *KvItem {
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

	kvs.size++
	return item
}

func (kvs *Kvs) Set(k string, v []byte) *KvItem {
	kvs.Del(k)
	return kvs.Append(k, v)
}

func (kvs *Kvs) SetString(k, v string) *KvItem {
	kvs.Del(k)
	return kvs.AppendString(k, v)
}

func (kvs *Kvs) Del(k string) {
	for i := 0; i < len(kvs.lst); i++ {
		item := &(kvs.lst[i])
		if item.invalid {
			continue
		}
		if string(item.Key) == k {
			item.makeInvalid()
			kvs.size--
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
	if kvs.size < 1 {
		return
	}

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
	if kvs.size < 1 {
		return
	}
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
	if kvs.size < 1 {
		return
	}
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
	kvs.size = 0
}

func (kvs *Kvs) LoadMap(m MultiValueMap) {
	for k, vl := range m {
		for _, v := range vl {
			kvs.AppendString(k, v)
		}
	}
}

func (kvs *Kvs) LoadKvs(o *Kvs) {
	o.EachItem(func(item *KvItem) bool {
		kvs.AppendBytes(item.Key, item.Val)
		return true
	})
}

func (kvs *Kvs) LoadAny(v interface{}) bool {
	switch tv := v.(type) {
	case url.Values, http.Header, MultiValueMap, map[string][]string:
		kvs.LoadMap(tv.(map[string][]string))
	case _KvsIterable:
		tv.EachItem(func(item *KvItem) bool {
			kvs.AppendBytes(item.Key, item.Val)
			return true
		})
	default:
		return false
	}
	return true
}
