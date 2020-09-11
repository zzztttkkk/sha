package sqls

import (
	"context"
	"strings"

	ci "github.com/zzztttkkk/suna/sqls/internal"
)

func (op *Operator) Exists(ctx context.Context, conditions ci.Sqlizer) bool {
	var v = -1
	q, a, e := Select("count(*)").From(op.TableName()).Where(conditions).ToSql()
	ExecuteSelect(ctx, &v, q, a, e)
	return v > 0
}

func (op *Operator) FetchMany(ctx context.Context, dist interface{}, conditions ci.Sqlizer, limit uint64, keys ...string) bool {
	columns := "*"
	if len(keys) > 0 {
		columns = strings.Join(keys, ",")
	}
	q, a, e := Select(columns).From(op.TableName()).Limit(limit).Where(conditions).ToSql()
	return ExecuteSelect(ctx, dist, q, a, e)
}

func (op *Operator) FetchOne(ctx context.Context, dist interface{}, conditions ci.Sqlizer, keys ...string) bool {
	return op.FetchMany(ctx, dist, conditions, 1, keys...)
}

func (op *Operator) Select(ctx context.Context, dist interface{}, builder *ci.SelectBuilder) bool {
	builder.FromIfEmpty(op.TableName())
	return ExecuteSelectBuilder(ctx, dist, builder)
}

func (op *Operator) Update(ctx context.Context, builder *ci.UpdateBuilder) int64 {
	builder.FromIfEmpty(op.TableName())

	n, e := ExecuteSql(ctx, builder).RowsAffected()
	if e != nil {
		panic(e)
	}
	return n
}

func (op *Operator) Delete(ctx context.Context, builder *ci.DeleteBuilder) int64 {
	builder.FromIfEmpty(op.TableName())

	n, e := ExecuteSql(ctx, builder).RowsAffected()
	if e != nil {
		panic(e)
	}
	return n
}

func (op *Operator) Insert(ctx context.Context, builder *ci.InsertBuilder) int64 {
	builder.IntoIfEmpty(op.TableName())

	if !IsPostgres() {
		n, e := ExecuteSql(ctx, builder).LastInsertId()
		if e != nil {
			panic(e)
		}
		return n
	}
	var lid int64
	query, args, err := builder.ToSql()
	ExecuteSelect(ctx, &lid, query, args, err)
	return lid
}
