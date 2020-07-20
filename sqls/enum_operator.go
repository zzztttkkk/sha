package sqls

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type EnumOperator struct {
	Operator
	cache *EnumCache
}

func (op *EnumOperator) NewEnumCache(seconds int64, constructor func() Enumer, initer func(context.Context, interface{}) error) *EnumCache {
	cache := &EnumCache{
		im:          map[int64]Enumer{},
		nm:          map[string]Enumer{},
		last:        0,
		expire:      seconds,
		op:          &op.Operator,
		constructor: constructor,
		initer:      initer,
		rwm:         sync.RWMutex{},
	}
	cache.load(context.Background())
	return cache
}

func (op *EnumOperator) Init(ele reflect.Value, constructor func() Enumer, initer func(context.Context, interface{}) error) {
	op.Operator.Init(ele)

	expire := config.GetIntOr(fmt.Sprintf("memcache.sqlenum.%s.expire", op.TableName()), -1)
	if expire < 1 {
		expire = config.GetIntOr("memcache.sqlenum.expire", 1800)
	}
	op.cache = op.NewEnumCache(expire, constructor, initer)
}

func (op *EnumOperator) Create(ctx context.Context, dict Dict) int64 {
	defer op.cache.doExpire()

	_, ok := dict["name"]
	if !ok {
		panic(fmt.Errorf("suna.sqls: enum key `name` is empty"))
	}
	dict["created"] = time.Now().Unix()
	return op.XCreate(ctx, dict)
}

func (op *EnumOperator) Delete(ctx context.Context, name string) bool {
	defer op.cache.doExpire()
	return op.XUpdate(
		ctx,
		Dict{"deleted": time.Now().Unix(), "name": fmt.Sprintf("!!!Deleted<%s>!!!", name)},
		"name=:name and deleted=0",
		Dict{"name": name},
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

func (op *EnumOperator) GetName(ctx context.Context, eid int64) string {
	e, ok := op.cache.GetById(ctx, eid)
	if !ok {
		return ""
	}
	return e.GetName()
}

func (op *EnumOperator) GetId(ctx context.Context, name string) int64 {
	e, ok := op.cache.GetByName(ctx, name)
	if !ok {
		return -1
	}
	return e.GetId()
}

func (op *EnumOperator) List(ctx context.Context) []Enumer {
	return op.cache.All(ctx)
}

func (op *EnumOperator) Expire() {
	op.cache.doExpire()
}
