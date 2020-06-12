package sqls

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/snow/ini"
	"reflect"
	"time"
)

type Enum struct {
	Model
	Name string `json:"name" ddl:"unique;notnull;L<30>"`
}

func (enum *Enum) GetId() int64 {
	return enum.Id
}

func (enum *Enum) GetName() string {
	return enum.Name
}

type EnumOperator struct {
	Operator
	cache *EnumCache
}

func (op *EnumOperator) Init(p reflect.Type, constructor func() Enumer) {
	op.Operator.Init(p)

	expire := ini.GetIntOr(fmt.Sprintf("memcache.sqlenum.%s.expire", op.ddl.tableName), -1)
	if expire < 1 {
		expire = ini.GetIntOr("memcache.sqlenum.expire", 1800)
	}
	op.cache = op.NewEnumCache(expire, constructor)
}

func (op *EnumOperator) Create(ctx context.Context, dict Dict) int64 {
	defer op.cache.doExpire()
	return op.SqlxCreate(ctx, dict)
}

func (op *EnumOperator) Delete(ctx context.Context, name string) bool {
	defer op.cache.doExpire()
	dict := Dict{}
	dict["deleted"] = time.Now().Unix()
	placeholder, values := dict.ForUpdate()
	values = append(values, name)

	return op.SqlxUpdate(
		ctx,
		fmt.Sprintf(
			"update %s set %s where name=? and deleted=0",
			op.ddl.tableName, placeholder,
		),
		values...,
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

func (op *EnumOperator) List(ctx context.Context) []Enumer {
	return op.cache.All(ctx)
}
