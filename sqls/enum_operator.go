package sqls

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/suna/sqls/builder"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
	"sync"
	"time"
)

type EnumOperator struct {
	Operator
	cache *EnumCache
}

func (op *EnumOperator) NewEnumCache(seconds int64, constructor func() EnumItem, afterScan func(context.Context, interface{}) error) *EnumCache {
	cache := &EnumCache{
		im:          map[int64]EnumItem{},
		nm:          map[string]EnumItem{},
		last:        0,
		expire:      seconds,
		op:          &op.Operator,
		constructor: constructor,
		afterScan:   afterScan,
		rwm:         sync.RWMutex{},
	}
	cache.load(context.Background())
	return cache
}

func (op *EnumOperator) Init(ele reflect.Value, constructor func() EnumItem, afterScan func(context.Context, interface{}) error) {
	op.Operator.Init(ele)

	expire := cfg.Sql.EnumCacheMaxage.Duration
	if expire < 1 {
		expire = time.Minute * 30
	}
	op.cache = op.NewEnumCache(int64(expire/time.Second), constructor, afterScan)
}

func (op *EnumOperator) Create(ctx context.Context, kvs *utils.Kvs) int64 {
	defer op.cache.doExpire()
	return op.XCreate(ctx, kvs)
}

func (op *EnumOperator) Delete(ctx context.Context, name string) bool {
	defer op.cache.doExpire()

	kvs := utils.AcquireKvs()
	defer kvs.Free()
	kvs.Set("deleted", time.Now().Unix())
	kvs.Set("name", fmt.Sprintf("Deleted<%s>", name))

	return op.XUpdate(
		ctx,
		kvs,
		builder.AndConditions().
			Eq(true, "name", name).
			Eq(true, "deleted", 0),
		1,
	) > 0
}

func (op *EnumOperator) ExistsById(ctx context.Context, eid int64) bool {
	_, ok := op.cache.GetById(ctx, eid)
	return ok
}

func (op *EnumOperator) ExistsByName(ctx context.Context, name string) bool {
	_, ok := op.cache.GetByName(ctx, name)
	return ok
}

func (op *EnumOperator) GetById(ctx context.Context, eid int64) (EnumItem, bool) {
	return op.cache.GetById(ctx, eid)
}

func (op *EnumOperator) GetByName(ctx context.Context, name string) (EnumItem, bool) {
	return op.cache.GetByName(ctx, name)
}

func (op *EnumOperator) GetNameById(ctx context.Context, eid int64) string {
	e, ok := op.cache.GetById(ctx, eid)
	if !ok {
		return ""
	}
	return e.GetName()
}

func (op *EnumOperator) GetIdByName(ctx context.Context, name string) int64 {
	e, ok := op.cache.GetByName(ctx, name)
	if !ok {
		return -1
	}
	return e.GetId()
}

func (op *EnumOperator) List(ctx context.Context) []EnumItem {
	return op.cache.All(ctx)
}

func (op *EnumOperator) Expire() {
	op.cache.doExpire()
}
