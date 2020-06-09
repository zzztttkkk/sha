package sqls

import (
	"context"
	"fmt"
	"golang.org/x/sync/singleflight"
	"sync"
	"time"
)

type Enum interface {
	GetId() int64
	GetName() string
}

type EnumCache struct {
	im          map[int64]interface{}
	nm          map[string]interface{}
	all         []Enum
	last        int64
	expire      int64
	op          *Operator
	constructor func() Enum
	rwm         sync.RWMutex
	sg          singleflight.Group
}

func (cache *EnumCache) doLoad(ctx context.Context) {
	_, _, _ = cache.sg.Do(
		"load",
		func() (interface{}, error) {
			cache.load(ctx)
			return nil, nil
		},
	)
}

func (cache *EnumCache) load(ctx context.Context) {
	cache.rwm.Lock()
	defer cache.rwm.Unlock()

	cache.all = make([]Enum, 0, len(cache.all))

	cache.op.SqlxStructScanRows(
		ctx,
		func() interface{} {
			obj := cache.constructor()
			cache.all = append(cache.all, obj)
			return obj
		},
		fmt.Sprintf(`select * from %s where deleted=0 and status>=0 order by id`, cache.op.TableName()),
	)

	for _, obj := range cache.all {
		cache.nm[obj.GetName()] = obj
		cache.im[obj.GetId()] = obj
	}
	cache.last = time.Now().Unix()
}

func (cache *EnumCache) isValid() bool {
	return time.Now().Unix()-cache.last <= cache.expire
}

func (cache *EnumCache) refresh(ctx context.Context) {
	cache.rwm.RLock()
	if cache.isValid() {
		cache.rwm.RUnlock()
		return
	}
	cache.rwm.RUnlock()

	cache.doLoad(ctx)
}

func (cache *EnumCache) GetById(ctx context.Context, id int64) (interface{}, bool) {
	cache.refresh(ctx)

	cache.rwm.RLock()
	defer cache.rwm.RUnlock()
	v, ok := cache.im[id]
	return v, ok
}

func (cache *EnumCache) GetByName(ctx context.Context, name string) (interface{}, bool) {
	cache.refresh(ctx)

	cache.rwm.RLock()
	defer cache.rwm.RUnlock()
	v, ok := cache.nm[name]
	return v, ok
}

func (cache *EnumCache) TraverseIdMap(ctx context.Context, visitor func(id int64, val interface{})) {
	cache.refresh(ctx)

	cache.rwm.RLock()
	defer cache.rwm.RUnlock()

	for k, v := range cache.im {
		visitor(k, v)
	}
}

func (cache *EnumCache) TraverseNameMap(ctx context.Context, visitor func(name string, val interface{})) {
	cache.refresh(ctx)

	cache.rwm.RLock()
	defer cache.rwm.RUnlock()

	for k, v := range cache.nm {
		visitor(k, v)
	}
}

func (cache *EnumCache) All(ctx context.Context) []Enum {
	cache.refresh(ctx)
	cache.rwm.RLock()
	defer cache.rwm.RUnlock()
	return cache.all
}
