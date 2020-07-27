package rbac

import (
	"context"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
)

type _PermOp struct {
	sqls.EnumOperator
	conflicts *sqls.Operator
}

var _PermissionOperator = &_PermOp{conflicts: &sqls.Operator{}}

func init() {
	lazier.RegisterWithPriority(
		func(kwargs utils.Kwargs) {
			_PermissionOperator.Init(
				reflect.ValueOf(_Permission{}),
				func() sqls.EnumItem { return &_Permission{} },
				nil,
			)
			_PermissionOperator.conflicts.Init(reflect.ValueOf(_PermConflict{}))
		},
		permTablePriority,
	)
}

func (op *_PermOp) Create(ctx context.Context, m utils.M) {
	defer LogOperator.Create(ctx, "perm.create", m)
	op.EnumOperator.Create(ctx, m)
}

func (op *_PermOp) Delete(ctx context.Context, name string) {
	defer LogOperator.Create(ctx, "perm.delete", utils.M{"name": name})
	op.EnumOperator.Delete(ctx, name)
}

func EnsurePermission(name, descp string) string {
	if _PermissionOperator.ExistsByName(context.Background(), name) {
		return name
	}
	_PermissionOperator.Create(context.Background(), utils.M{"name": name, "descp": descp})
	return name
}
