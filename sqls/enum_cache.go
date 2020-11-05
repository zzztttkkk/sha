package sqls

import (
	"context"
	"sync"
	"time"
)

type _EnumCache struct {
	idMap       map[int64]EnumItem
	nameMap     map[string]EnumItem
	temps       []EnumItem
	lastChange  int64
	expires     int64
	op          *Operator
	constructor func() EnumItem
	afterLoad   func(ctx context.Context, items []EnumItem)
	rwm         sync.RWMutex
}

func (cache *_EnumCache) load(ctx context.Context) {
	cache.rwm.Lock()
	defer cache.rwm.Unlock()

	var temp interface{}

	ExecuteCustomScan(
		ctx,
		NewStructScanner(
			&temp,
			func() {
				temp = cache.constructor()
			},
			func() {
				cache.temps = append(cache.temps, temp.(EnumItem))
			},
		),
		Select("*").
			From(cache.op.TableName()).
			Where("status>=0 and deleted=0").
			OrderBy("id"),
	)

	for _, obj := range cache.temps {
		cache.nameMap[obj.GetName()] = obj
		cache.idMap[obj.GetId()] = obj
	}
	cache.lastChange = time.Now().Unix()

	if cache.afterLoad != nil {
		cache.afterLoad(ctx, cache.temps)
	}
}

func (cache *_EnumCache) refresh(ctx context.Context) {
	cache.rwm.RLock()
	if time.Now().Unix()-cache.lastChange <= cache.expires {
		cache.rwm.RUnlock()
		return
	}
	cache.rwm.RUnlock()

	cache.load(ctx)
}

func (cache *_EnumCache) doExpire() {
	cache.rwm.Lock()
	defer cache.rwm.Unlock()
	cache.lastChange = 0
}

func (cache *_EnumCache) GetById(ctx context.Context, id int64) (EnumItem, bool) {
	cache.refresh(ctx)

	cache.rwm.RLock()
	defer cache.rwm.RUnlock()

	v, ok := cache.idMap[id]
	return v, ok
}

func (cache *_EnumCache) GetByName(ctx context.Context, name string) (EnumItem, bool) {
	cache.refresh(ctx)

	cache.rwm.RLock()
	defer cache.rwm.RUnlock()

	v, ok := cache.nameMap[name]
	return v, ok
}

func (cache *_EnumCache) TraverseIdMap(ctx context.Context, visitor func(id int64, val interface{})) {
	cache.refresh(ctx)

	cache.rwm.RLock()
	defer cache.rwm.RUnlock()

	for k, v := range cache.idMap {
		visitor(k, v)
	}
}

func (cache *_EnumCache) TraverseNameMap(ctx context.Context, visitor func(name string, val interface{})) {
	cache.refresh(ctx)

	cache.rwm.RLock()
	defer cache.rwm.RUnlock()

	for k, v := range cache.nameMap {
		visitor(k, v)
	}
}

func (cache *_EnumCache) All(ctx context.Context) []EnumItem {
	cache.refresh(ctx)

	cache.rwm.RLock()
	defer cache.rwm.RUnlock()

	return cache.temps
}
