package sqls

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
	"sync"
	"time"
)

type EnumOperator struct {
	Operator
	cache *EnumCache
}

func (op *EnumOperator) NewEnumCache(seconds int64, constructor func() EnumItem, initer func(context.Context, interface{}) error) *EnumCache {
	cache := &EnumCache{
		im:          map[int64]EnumItem{},
		nm:          map[string]EnumItem{},
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

func (op *EnumOperator) Init(ele reflect.Value, constructor func() EnumItem, initer func(context.Context, interface{}) error) {
	op.Operator.Init(ele)

	expire := config.GetIntOr(fmt.Sprintf("cache.sqlenum.%s.expire", op.TableName()), -1)
	if expire < 1 {
		expire = config.GetIntOr("cache.sqlenum.expire", 1800)
	}
	op.cache = op.NewEnumCache(expire, constructor, initer)
}

func (op *EnumOperator) Create(ctx context.Context, dict utils.M) int64 {
	defer op.cache.doExpire()

	_, ok := dict["name"]
	if !ok {
		panic(fmt.Errorf("suna.sqlu: enum key `name` is empty"))
	}
	dict["created"] = time.Now().Unix()
	return op.XCreate(ctx, dict)
}

func (op *EnumOperator) Delete(ctx context.Context, name string) bool {
	defer op.cache.doExpire()
	return op.XUpdate(
		ctx,
		utils.M{"deleted": time.Now().Unix(), "name": fmt.Sprintf("Deleted<%s>", name)},
		"name=:name and deleted=0",
		utils.M{"name": name},
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
