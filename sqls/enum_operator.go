package sqls

import (
	"context"
	"fmt"
	"time"
)

type EnumOperator struct {
	Operator
	cache *_EnumCache
}

func (op *EnumOperator) newEnumCache(
	seconds int64,
	constructor func() EnumItem,
	afterLoad func(ctx context.Context, items []EnumItem),
) *_EnumCache {
	cache := &_EnumCache{
		idMap:       map[int64]EnumItem{},
		nameMap:     map[string]EnumItem{},
		lastChange:  0,
		expires:     seconds,
		op:          &op.Operator,
		constructor: constructor,
		afterLoad:   afterLoad,
	}
	cache.load(context.Background())
	return cache
}

func (op *EnumOperator) Init(
	ele interface{},
	constructor func() EnumItem,
	afterLoad func(ctx context.Context, items []EnumItem),
) {
	op.Operator.Init(ele)

	expire := cfg.Sql.EnumCacheMaxAge.Duration
	if expire < 1 {
		expire = time.Minute * 30
	}
	op.cache = op.newEnumCache(int64(expire/time.Second), constructor, afterLoad)
}

func (op *EnumOperator) Create(ctx context.Context, name, descp string) int64 {
	defer op.cache.doExpire()
	builder := Insert("name,descp,created").Values(name, descp, time.Now().Unix())
	if isPostgres {
		builder = builder.Returning("id")
	}
	return op.ExecInsert(ctx, builder)
}

func (op *EnumOperator) Delete(ctx context.Context, name string) bool {
	defer op.cache.doExpire()

	builder := Update().
		Set("name", fmt.Sprintf("Deleted<%s>", name)).
		Set("deleted", time.Now().Unix()).
		Where(STR("name=? and deleted=0", name))
	return op.ExecUpdate(ctx, builder) > 0
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
