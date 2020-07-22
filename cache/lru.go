package cache

import (
	"github.com/golang/groupcache/lru"
	"sync"
)

type Lru struct {
	mutex sync.RWMutex
	raw   *lru.Cache
}

func NewLru(maxEntries int) *Lru { return &Lru{raw: lru.New(maxEntries)} }

func (cache *Lru) Add(key interface{}, val interface{}) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.raw.Add(key, val)
}

func (cache *Lru) Get(key interface{}) (interface{}, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.raw.Get(key)
}

func (cache *Lru) Remove(key interface{}) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.raw.Remove(key)
}

func (cache *Lru) RemoveOldest() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.raw.RemoveOldest()
}

func (cache *Lru) Len() int {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.raw.Len()
}

func (cache *Lru) Clear() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.raw.Clear()
}
