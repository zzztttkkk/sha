package rbac

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/sqls/builder"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
	"time"
)

type logOpT struct {
	sqls.Operator
}

var LogOperator = &logOpT{}

func init() {
	lazier.RegisterWithPriority(
		func(kwargs utils.Kwargs) {
			LogOperator.Init(reflect.ValueOf(logT{}))
		},
		initPriority.Incr(),
	)
}

func (op *logOpT) Create(ctx context.Context, name string, info utils.M) int64 {
	v := recover()
	if v != nil {
		panic(v)
	}

	user := sqls.TxOperator(ctx)
	if user == nil {
		panic(output.HttpErrors[fasthttp.StatusUnauthorized])
	}

	kvs := utils.AcquireKvs()
	defer kvs.Free()

	kvs.Set("created", time.Now())
	kvs.Set("name", name)
	kvs.Set("operator", user.GetId())
	kvs.Set("info", utils.JsonObject(info))
	return op.XCreate(ctx, kvs)
}

func (op *logOpT) List(
	ctx context.Context,
	begin, end int64,
	names []string, uids []int64,
	asc bool,
	cursor int64,
	limit int,
) (lst []logT) {
	conditions := builder.NewConditions(builder.AND)
	conditions.Gte(begin > 0, "created", begin)
	conditions.Lte(end > 0, "created", end)
	conditions.Eq(len(names) > 0, "name", names)
	conditions.Eq(len(uids) > 0, "operator", uids)
	if asc {
		conditions.Gt(cursor > 0, "id", cursor)
	} else {
		conditions.Lt(cursor > 0, "id", cursor)
	}

	sb := builder.NewSelect("*").From(op.TableName()).Where(conditions).Limit(uint64(limit))
	if asc {
		sb.OrderBy("id")
	} else {
		sb.OrderBy("id desc")
	}

	op.XSelect(ctx, &lst, sb)
	return
}
