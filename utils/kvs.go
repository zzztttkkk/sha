package utils

import "sync"

type _KvNode struct {
	isValid bool
	k       string
	v       interface{}
	next    *_KvNode
}

var nodePool = sync.Pool{New: func() interface{} { return &_KvNode{} }}

func getNode(k string, v interface{}) *_KvNode {
	node := nodePool.Get().(*_KvNode)
	node.k = k
	node.v = v
	node.isValid = true
	return node
}

func putNode(node *_KvNode) {
	node.k = ""
	node.v = nil
	node.next = nil
	node.isValid = false
	nodePool.Put(node)
}

type Kvs struct {
	head *_KvNode
	tail *_KvNode
}

func (kvs *Kvs) Append(k string, v interface{}) {
	node := getNode(k, v)

	if kvs.tail == nil {
		kvs.tail = node
		kvs.head = node
		return
	}

	kvs.tail.next = node
	kvs.tail = node
}

func (kvs *Kvs) Set(k string, v interface{}) {
	c := kvs.head
	for c != nil {
		if c.k == k {
			c.v = v
			if !c.isValid {
				c.isValid = true
			}
			return
		}
		c = c.next
	}
	kvs.Append(k, v)
}

func (kvs *Kvs) Remove(k string) {
	c := kvs.head
	for c != nil {
		if c.k == k {
			c.isValid = false
			c.v = nil
			return
		}
	}
}

func (kvs *Kvs) EachKey(fn func(string)) {
	c := kvs.head
	for c != nil {
		if c.isValid {
			fn(c.k)
		}
		c = c.next
	}
}

func (kvs *Kvs) EachValue(fn func(interface{})) {
	c := kvs.head
	for c != nil {
		if c.isValid {
			fn(c.v)
		}
		c = c.next
	}
}

func (kvs *Kvs) EachNode(fn func(string, interface{})) {
	c := kvs.head
	for c != nil {
		if c.isValid {
			fn(c.k, c.v)
		}
		c = c.next
	}
}

var dictPool = sync.Pool{New: func() interface{} { return &Kvs{} }}

func AcquireKvs() *Kvs {
	return dictPool.Get().(*Kvs)
}

func (kvs *Kvs) Free() {
	c := kvs.head
	for c != nil {
		n := c.next
		putNode(c)
		c = n
	}

	kvs.head = nil
	kvs.tail = nil
	dictPool.Put(kvs)
}
