package rbac

import (
	"context"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
)

type _PermOp struct {
	sqls.EnumOperator
}

var _PermissionOperator = &_PermOp{}

func init() {
	lazier.RegisterWithPriority(
		func(kwargs utils.Kwargs) {
			_PermissionOperator.Init(
				reflect.ValueOf(_Permission{}),
				func() sqls.EnumItem { return &_Permission{} },
				nil,
			)
		},
		permTablePriority,
	)
}

func (op *_PermOp) Create(ctx context.Context, m utils.M) {
	defer LogOperator.Create(ctx, "perm.create", m)
	op.create(ctx, m)
}

func (op *_PermOp) create(ctx context.Context, m utils.M) {
	op.EnumOperator.Create(ctx, m)
}

func (op *_PermOp) Delete(ctx context.Context, name string) {
	defer LogOperator.Create(ctx, "perm.delete", utils.M{"name": name})
	op.EnumOperator.Delete(ctx, name)
}

func (op *_PermOp) List(ctx context.Context) (lst []*_Permission) {
	for _, enum := range op.EnumOperator.List(ctx) {
		lst = append(lst, enum.(*_Permission))
	}
	return
}

func EnsurePermission(name, descp string) string {
	if _PermissionOperator.ExistsByName(context.Background(), name) {
		return name
	}
	_PermissionOperator.create(context.Background(), utils.M{"name": name, "descp": descp})
	return name
}
