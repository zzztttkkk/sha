package utils

import "sync"

type _KvNode struct {
	isValid bool
	key     string
	value   interface{}
}

type Kvs struct {
	data []_KvNode
}

func (kvs *Kvs) Append(k string, v interface{}) *Kvs {
	kvs.data = append(kvs.data, _KvNode{key: k, value: v, isValid: true})
	return kvs
}

func (kvs *Kvs) Set(k string, v interface{}) *Kvs {
	for i := 0; i < len(kvs.data); i++ {
		node := &kvs.data[i]
		if !node.isValid {
			continue
		}
		if node.key == k {
			node.value = v
			return kvs
		}
	}
	return kvs.Append(k, v)
}

func (kvs *Kvs) Get(k string) (interface{}, bool) {
	for i := 0; i < len(kvs.data); i++ {
		node := &kvs.data[i]
		if !node.isValid {
			continue
		}
		if node.key == k {
			return node.value, true
		}
	}
	return nil, false
}

func (kvs *Kvs) ToMap() (m map[string]interface{}) {
	for i := 0; i < len(kvs.data); i++ {
		node := &kvs.data[i]
		if !node.isValid {
			continue
		}
		if m == nil {
			m = make(map[string]interface{})
		}
		m[node.key] = node.value
	}
	return
}

func (kvs *Kvs) FromMap(m map[string]interface{}) *Kvs {
	for k, v := range m {
		kvs.Append(k, v)
	}
	return kvs
}

func (kvs *Kvs) Remove(k string) (v interface{}, ok bool) {
	for i := 0; i < len(kvs.data); i++ {
		node := &kvs.data[i]
		if !node.isValid {
			continue
		}
		if node.key == k {
			node.isValid = false
			return node.value, true
		}
	}
	return nil, false
}

func (kvs *Kvs) EachKey(fn func(string)) {
	for i := 0; i < len(kvs.data); i++ {
		node := &kvs.data[i]
		if !node.isValid {
			continue
		}
		fn(node.key)
	}
}

func (kvs *Kvs) EachValue(fn func(interface{})) {
	for i := 0; i < len(kvs.data); i++ {
		node := &kvs.data[i]
		if !node.isValid {
			continue
		}
		fn(node.value)
	}
}

func (kvs *Kvs) EachNode(fn func(string, interface{})) {
	for i := 0; i < len(kvs.data); i++ {
		node := &kvs.data[i]
		if !node.isValid {
			continue
		}
		fn(node.key, node.value)
	}
}

var kvsPool = sync.Pool{New: func() interface{} { return &Kvs{} }}

func AcquireKvs() *Kvs {
	return kvsPool.Get().(*Kvs)
}

func (kvs *Kvs) Free() {
	kvs.data = kvs.data[:0]
	kvsPool.Put(kvs)
}
