package rbac

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/sqls/builder"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
)

type roleOpT struct {
	sqls.EnumOperator
	perms     *sqls.Operator
	inherits  *sqls.Operator
	conflicts *sqls.Operator
}

var _RoleOperator = &roleOpT{
	perms:    &sqls.Operator{},
	inherits: &sqls.Operator{},
}

func init() {
	lazier.RegisterWithPriority(
		func(kwargs utils.Kwargs) {
			_RoleOperator.perms.Init(reflect.ValueOf(roleWithPermT{}))
			_RoleOperator.inherits.Init(reflect.ValueOf(roleInheritanceT{}))

			_RoleOperator.EnumOperator.Init(
				reflect.ValueOf(roleT{}),
				func() sqls.EnumItem { return &roleT{} },
				func(ctx context.Context, i interface{}) error {
					role := i.(*roleT)
					_RoleOperator.getAllPerms(ctx, role)
					_RoleOperator.getAllBasedRoles(ctx, role)
					return nil
				},
			)
		},
		permTablePriority.Incr(),
	)
}

func (op *roleOpT) changePerm(ctx context.Context, roleName, permName string, mt modifyType) error {
	permId := _PermissionOperator.GetIdByName(ctx, permName)
	if permId < 1 {
		return output.HttpErrors[fasthttp.StatusNotFound]
	}
	roleId := _RoleOperator.GetIdByName(ctx, roleName)
	if roleId < 1 {
		return output.HttpErrors[fasthttp.StatusNotFound]
	}

	defer LogOperator.Create(
		ctx,
		"role.changePerm",
		utils.M{
			"perm":   fmt.Sprintf("%d:%s", permId, permName),
			"role":   fmt.Sprintf("%d:%s", roleId, roleName),
			"modify": mt.String(),
		},
	)

	cond := builder.AndConditions().
		Eq(true, "role", roleId).
		Eq(true, "perm", permId)

	spb := builder.NewSelect("id").From(op.perms.TableName()).Where(cond)

	var _id int64
	op.perms.XSelect(ctx, &_id, spb)
	if _id < 1 {
		if mt == _Add {
			return nil
		}

		kvs := utils.AcquireKvs()
		defer kvs.Free()
		kvs.Set("perm", permId)
		kvs.Set("role", roleId)

		op.perms.XCreate(ctx, kvs)
		return nil
	}

	if mt == _Add {
		return nil
	}

	q, args, err := builder.NewDelete().From(op.perms.TableName()).Where(cond).Limit(1).ToSql()
	if err != nil {
		panic(err)
	}
	op.perms.XExecute(ctx, q, args...)
	return nil
}

func (op *roleOpT) getAllPerms(ctx context.Context, role *roleT) {
	sb := builder.NewSelect("perm").Prefix("distinct").From(op.perms.TableName()).
		Where("role=?", role.Id)

	op.perms.XSelect(ctx, &role.Permissions, sb)
}

func (op *roleOpT) changeInherits(ctx context.Context, roleName, basedRoleName string, mt modifyType) error {
	roleId := _RoleOperator.GetIdByName(ctx, roleName)
	if roleId < 1 {
		return output.HttpErrors[fasthttp.StatusNotFound]
	}

	basedRoleId := _RoleOperator.GetIdByName(ctx, roleName)
	if basedRoleId < 1 {
		return output.HttpErrors[fasthttp.StatusNotFound]
	}

	defer LogOperator.Create(
		ctx,
		"role.changeInherits",
		utils.M{
			"role":   fmt.Sprintf("%d:%s", roleId, roleName),
			"based":  fmt.Sprintf("%d:%s", basedRoleId, basedRoleName),
			"modify": mt.String(),
		},
	)

	cond := builder.AndConditions().
		Eq(true, "role", roleId).
		Eq(true, "based", basedRoleId)

	var _id int64
	op.inherits.XSelect(ctx, &_id, builder.NewSelect("based").From(op.inherits.TableName()).Where(cond))
	if _id < 1 {
		if mt == _Add {
			return nil
		}
		kvs := utils.AcquireKvs()
		defer kvs.Free()
		kvs.Set("role", roleId)
		kvs.Set("based", basedRoleId)
		op.perms.XCreate(ctx, kvs)
		return nil
	}

	if mt == _Add {
		return nil
	}

	q, args, err := builder.NewDelete().From(op.inherits.TableName()).Where(cond).Limit(1).ToSql()
	if err != nil {
		panic(err)
	}
	op.inherits.XExecute(ctx, q, args...)
	return nil
}

func (op *roleOpT) getAllBasedRoles(ctx context.Context, role *roleT) {
	sb := builder.NewSelect("based").Prefix("distinct").From(op.inherits.TableName()).
		Where("role=?", role.Id)
	op.inherits.XSelect(ctx, &role.Based, sb)
}

func (op *roleOpT) List(ctx context.Context) (lst []*roleT) {
	for _, enum := range op.EnumOperator.List(ctx) {
		lst = append(lst, enum.(*roleT))
	}
	return
}
