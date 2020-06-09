package sqls

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type Operator struct {
	p   reflect.Type
	ddl *_DdlParser
}

func (op *Operator) Init(p reflect.Type) {
	op.p = p
	op.ddl = newDdlParser(p)
}

func (op *Operator) SqlsTableCreate() {
	master.MustExec(TableDefinition(op.p))
}

func (op *Operator) SqlsModelExists(ctx context.Context, key string, val interface{}) bool {
	c := -1
	op.SqlxFetchOne(
		ctx,
		&c,
		fmt.Sprintf(`select count(id) from %s where %s=? and deleted=0 and status>=0`, op.ddl.tableName, key),
		val,
	)
	return c > 0
}

func (op *Operator) TableName() string {
	return op.ddl.tableName
}

func (op *Operator) NewEnumCache(seconds int64, constructor func() Enum) *EnumCache {
	cache := &EnumCache{
		im:          map[int64]interface{}{},
		nm:          map[string]interface{}{},
		last:        0,
		expire:      seconds,
		op:          op,
		constructor: constructor,
		rwm:         sync.RWMutex{},
	}
	cache.load(context.Background())
	return cache
}
