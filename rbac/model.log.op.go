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
	dig.Append(
		func() _DigLogTableInited {
			LogOperator.Init(logT{})
			return _DigLogTableInited(0)
		},
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

	builder := sqls.Insert("name,operator,info,created").
		Values(name, user.GetId(), jsonx.Object(info), time.Now().Unix())
	if sqls.IsPostgres() {
		builder = builder.Returning("id")
	}
	return op.ExecInsert(ctx, builder)
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
	conditions.GteIf(begin > 0, "created", begin).
		LteIf(end > 0, "created", end).
		EqIf(len(names) > 0, "name", names).
		EqIf(len(uids) > 0, "operator", uids)
	if asc {
		conditions.GtIf(cursor > 0, "id", cursor)
	} else {
		conditions.LtIf(cursor > 0, "id", cursor)
	}

	if limit < 1 {
		limit = 1
	}
	op.FetchMany(ctx, &lst, conditions, uint64(limit))
	return
}
