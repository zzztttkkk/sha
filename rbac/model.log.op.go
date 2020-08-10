package rbac

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
	"strings"
	"time"
)

type _LogOp struct {
	sqls.Operator
}

var LogOperator = &_LogOp{}

func init() {
	lazier.RegisterWithPriority(
		func(kwargs utils.Kwargs) {
			LogOperator.Init(reflect.ValueOf(Log{}))
		},
		initPriority.Incr(),
	)
}

func (op *_LogOp) Create(ctx context.Context, name string, info utils.M) int64 {
	v := recover()
	if v != nil {
		panic(v)
	}

	user := sqls.TxOperator(ctx)
	if user == nil {
		panic(output.HttpErrors[fasthttp.StatusUnauthorized])
	}
	return op.XCreate(
		ctx,
		utils.M{
			"created":  time.Now().Unix(),
			"name":     name,
			"operator": user.GetId(),
			"info":     utils.JsonObject(info),
		},
	)
}

const (
	LogOrderByCreated = iota + 1
	LogOrderByName
	LogOrderByUid
)

func (op *_LogOp) List(
	ctx context.Context,
	begin, end int64, names []string,
	uids []int64,
	order int,
	asc bool,
	cursor interface{},
	limit int,
) (lst []*Log) {
	var filters []string
	var args []interface{}

	// time
	if end < begin {
		begin, end = end, begin
	}
	if begin > 0 {
		filters = append(filters, fmt.Sprintf("created>=%d", begin))
	}
	if end > 0 {
		filters = append(filters, fmt.Sprintf("created<=%d", end))
	}

	// names
	if len(names) > 0 {
		filters = append(filters, "name in ?")
		args = append(args, names)
	}

	// uids
	if len(uids) > 0 {
		filters = append(filters, "operator in ?")
		args = append(args, uids)
	}

	// order
	orderKey := "created"
	switch order {
	case LogOrderByName:
		orderKey = "name"
	case LogOrderByUid:
		orderKey = "operator"
	}

	filters = append(filters, fmt.Sprintf("%s>?", orderKey))
	args = append(args, cursor)

	q := fmt.Sprintf("select * from %s", op.TableName())
	if len(filters) > 0 {
		q = q + strings.Join(filters, " and ")
	}
	if asc {
		q += fmt.Sprintf(" order by %s", orderKey)
	} else {
		q += fmt.Sprintf(" order by %s desc", orderKey)
	}

	q += "limit ?"
	args = append(args, limit)

	op.XQsn(
		ctx,
		func() interface{} {
			log := &Log{}
			lst = append(lst, log)
			return log
		},
		nil,
		q,
		args...,
	)
	return
}
