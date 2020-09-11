package rbac

import (
	"context"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/jsonx"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
)

type logOpT struct {
	sqls.Operator
}

var LogOperator = &logOpT{}

func init() {
	lazier.RegisterWithPriority(
		func(kwargs utils.Kwargs) {
			LogOperator.Init(logT{})
		},
		initPriority.Incr(),
	)
}

func (op *logOpT) Create(ctx context.Context, name string, info utils.M) int64 {
	v := recover()
	if v != nil {
		panic(v)
	}

	user := sqls.TxUser(ctx)
	if user == nil {
		panic(output.HttpErrors[fasthttp.StatusUnauthorized])
	}

	builder := sqls.Insert(op.TableName()).
		Columns("name,operator,info,created").
		Values(name, user.GetId(), jsonx.Object(info), time.Now().Unix())
	if sqls.IsPostgres() {
		builder = builder.Returning("id")
	}
	return op.Insert(ctx, builder)
}

func (op *logOpT) List(
	ctx context.Context,
	begin, end int64,
	names []string, uids []int64,
	asc bool,
	cursor int64,
	limit int64,
) (lst []logT) {
	conditions := sqls.AND()
	conditions.Gte(begin > 0, "created", begin).
		Lte(end > 0, "created", end).
		Eq(len(names) > 0, "name", names).
		Eq(len(uids) > 0, "operator", uids)
	if asc {
		conditions.Gt(cursor > 0, "id", cursor)
	} else {
		conditions.Lt(cursor > 0, "id", cursor)
	}

	if limit < 1 {
		limit = 1
	}
	op.FetchMany(ctx, &lst, conditions, uint64(limit))
	return
}
