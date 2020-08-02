package rbac

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
	"time"
)

type _LogOp struct {
	sqls.Operator
}

var LogOperator = &_LogOp{}

func (op *_LogOp) Create(ctx context.Context, name string, info utils.M) int64 {
	v := recover()
	if v != nil {
		panic(v)
	}

	user := getUserFromCtx(ctx)
	if user == nil {
		panic(output.HttpErrors[fasthttp.StatusUnauthorized])
	}
	return op.XCreate(
		ctx,
		utils.M{
			"created":  time.Now().Unix(),
			"name":     name,
			"operator": user.GetId(),
			"info":     sqls.JsonObject(info),
		},
	)
}
