package sqls

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type EnumCache struct {
	im          map[int64]EnumItem
	nm          map[string]EnumItem
	all         []EnumItem
	last        int64
	expire      int64
	op          *Operator
	constructor func() EnumItem
	afterScan   func(context.Context, interface{}) error
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

	cache.all = make([]EnumItem, 0, len(cache.all))

	cache.op.XQsn(
		ctx,
		func() interface{} {
			obj := cache.constructor()
			cache.all = append(cache.all, obj)
			return obj
		},
		cache.afterScan,
		fmt.Sprintf(`select * from %s where deleted=0 and status>=0 order by id`, cache.op.TableName()),
	)

	for _, obj := range cache.all {
		cache.nm[obj.GetName()] = obj
		cache.im[obj.GetId()] = obj
	}
	cache.last = time.Now().Unix()
}

func (cache *EnumCache) refresh(ctx context.Context) {
	cache.rwm.RLock()

	if time.Now().Unix()-cache.last <= cache.expire {
		cache.rwm.RUnlock()
		return
	}

	cache.rwm.RUnlock()

	cache.doLoad(ctx)
}

func (cache *EnumCache) doExpire() {
	cache.rwm.Lock()
	defer cache.rwm.Unlock()
	cache.last = 0
}

func (cache *EnumCache) GetById(ctx context.Context, id int64) (EnumItem, bool) {
	cache.refresh(ctx)

	cache.rwm.RLock()
	defer cache.rwm.RUnlock()

	v, ok := cache.im[id]
	return v, ok
}

func (cache *EnumCache) GetByName(ctx context.Context, name string) (EnumItem, bool) {
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

func (cache *EnumCache) All(ctx context.Context) []EnumItem {
	cache.refresh(ctx)

	cache.rwm.RLock()
	defer cache.rwm.RUnlock()

	return cache.all
}
