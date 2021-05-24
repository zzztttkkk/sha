package groupcache

import (
	"fmt"
	"testing"
)

func Test_LRU(t *testing.T) {
	cache := NewLRUCache(100)

	for i := 0; i < 10000; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), i, 0)
	}

	fmt.Println(cache.Get("key-9909"))
	fmt.Println(cache.lst.Front().Value, cache.lst.Len())
}
