package groupcache

import (
	"container/list"
	"context"
	"github.com/zzztttkkk/sha/utils"
	"sync"
	"time"
)

type _LRUItem struct {
	key    []byte
	val    []byte
	expire int64
}

type LRUCacheStorage struct {
	sync.RWMutex
	lst   list.List
	m     map[string]*list.Element
	items []*_LRUItem
	cap   int
}

func (s *LRUCacheStorage) getEle(key string) *list.Element {
	s.RLock()
	defer s.RUnlock()

	ele := s.m[key]
	if ele != nil {
		s.lst.MoveToFront(ele)
		return ele
	}
	return nil
}

func (s *LRUCacheStorage) Get(_ context.Context, key string) ([]byte, bool) {
	ele := s.getEle(key)
	if ele == nil {
		return nil, false
	}
	item := ele.Value.(*_LRUItem)
	if item.expire > 0 && item.expire < time.Now().UnixNano() {
		s.Lock()
		s.del(ele)
		s.Unlock()
		return nil, false
	}
	return item.val, true
}

func (s *LRUCacheStorage) del(element *list.Element) {
	item := element.Value.(*_LRUItem)
	delete(s.m, utils.S(item.key))
	s.lst.Remove(element)
	item.val = nil
	item.expire = 0
	item.key = item.key[:0]
	s.items = append(s.items, item)
}

func (s *LRUCacheStorage) Set(_ context.Context, key string, val []byte, expire time.Duration) {
	s.Lock()
	defer s.Unlock()

	ele := s.m[key]
	if ele != nil {
		s.lst.MoveToFront(ele)
		item := ele.Value.(*_LRUItem)
		item.val = val
		if expire > 0 {
			item.expire = time.Now().UnixNano() + int64(expire)
		} else {
			item.expire = 0
		}
		return
	}

	var item *_LRUItem
	if len(s.items) > 0 {
		item = s.items[len(s.items)-1]
		s.items = s.items[:len(s.items)-1]
	} else {
		item = &_LRUItem{}
	}

	item.val = val
	item.key = append(item.key, key...)
	if expire > 0 {
		item.expire = time.Now().UnixNano() + int64(expire)
	}
	s.lst.PushFront(item)
	ele = s.lst.Front()
	s.m[key] = ele
	if s.lst.Len() > s.cap {
		s.del(s.lst.Back())
	}
}

func (s *LRUCacheStorage) Del(_ context.Context, keys ...string) {
	s.Lock()
	defer s.Unlock()

	for _, key := range keys {
		ele := s.m[key]
		if ele == nil {
			continue
		}
		s.del(ele)
	}
}

func NewLRUCache(cap int) *LRUCacheStorage {
	return &LRUCacheStorage{cap: cap, m: map[string]*list.Element{}}
}
