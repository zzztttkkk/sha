package rbac

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
	"time"
)

type permOpT struct {
	sqls.EnumOperator
}

var _PermissionOperator = &permOpT{}

func init() {
	lazier.RegisterWithPriority(
		func(kwargs utils.Kwargs) {
			_PermissionOperator.Init(
				Permission{},
				func() sqls.EnumItem { return &Permission{} },
				nil,
			)
		},
		permTablePriority,
	)
}

func (op *permOpT) Create(ctx context.Context, m utils.M) {
	defer LogOperator.Create(ctx, "perm.create", m)
	op.create(ctx, m)
}

func (op *permOpT) create(ctx context.Context, m utils.M) {
	kvs := utils.AcquireKvs()
	defer kvs.Free()
	kvs.FromMap(m)
	op.EnumOperator.Create(ctx, kvs)
}

func (op *permOpT) Delete(ctx context.Context, name string) {
	defer LogOperator.Create(ctx, "perm.delete", utils.M{"Name": name})
	op.EnumOperator.Delete(ctx, name)
}

func (op *permOpT) List(ctx context.Context) (lst []*Permission) {
	for _, enum := range op.EnumOperator.List(ctx) {
		lst = append(lst, enum.(*Permission))
	}
	return
}

func EnsurePermission(name, descp string) string {
	if _PermissionOperator.ExistsByName(context.Background(), name) {
		return name
	}

	tcx, committer := sqls.Tx(context.Background())
	defer committer()
	_PermissionOperator.create(
		tcx,
		utils.M{
			"name":    name,
			"descp":   fmt.Sprintf("%s; created by `EnsurePermission`", descp),
			"created": time.Now().Unix(),
		},
	)
	return name
}
