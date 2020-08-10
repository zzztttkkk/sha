package rbac

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
	"reflect"
)

type _RoleOp struct {
	sqls.EnumOperator
	perms     *sqls.Operator
	inherits  *sqls.Operator
	conflicts *sqls.Operator
}

var _RoleOperator = &_RoleOp{
	perms:    &sqls.Operator{},
	inherits: &sqls.Operator{},
}

func init() {
	lazier.RegisterWithPriority(
		func(kwargs utils.Kwargs) {
			_RoleOperator.perms.Init(reflect.ValueOf(_RoleWithPerm{}))
			_RoleOperator.inherits.Init(reflect.ValueOf(_RoleInheritance{}))

			_RoleOperator.EnumOperator.Init(
				reflect.ValueOf(_Role{}),
				func() sqls.EnumItem { return &_Role{} },
				func(ctx context.Context, i interface{}) error {
					role := i.(*_Role)
					_RoleOperator.getAllPerms(ctx, role)
					_RoleOperator.getAllBasedRoles(ctx, role)
					return nil
				},
			)
		},
		permTablePriority.Incr(),
	)
}

func (op *_RoleOp) changePerm(ctx context.Context, roleName, permName string, mt modifyType) error {
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

	q, vl := op.BindNamed(
		fmt.Sprintf("select perm from %s where role=:rid and perm=:pid", op.perms.TableName()),
		utils.M{"rid": roleId, "pid": permId},
	)
	var _id int64
	op.perms.XQ11(ctx, &_id, q, vl...)
	if _id < 1 {
		if mt == Add {
			return nil
		}
		op.perms.XCreate(
			ctx,
			utils.M{
				"perm": permId,
				"role": roleId,
			},
		)
		return nil
	}

	if mt == Add {
		return nil
	}

	q, vl = op.BindNamed("delete from %s where role=:rid and perm=:pid", utils.M{"rid": roleId, "pid": permId})
	op.perms.XExecute(ctx, q, vl...)
	return nil
}

func (op *_RoleOp) getAllPerms(ctx context.Context, role *_Role) {
	op.perms.XQ1n(ctx, &role.Permissions, fmt.Sprintf(`select distinct perm from %s where role=?`, op.perms.TableName()), role.Id)
}

func (op *_RoleOp) changeInherits(ctx context.Context, roleName, basedRoleName string, mt modifyType) error {
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

	q, vl := op.BindNamed(
		fmt.Sprintf("select role from %s where role=:rid and based=:brid", op.inherits.TableName()),
		utils.M{"rid": roleId, "brid": basedRoleId},
	)
	var _id int64
	op.inherits.XQ11(ctx, &_id, q, vl...)
	if _id < 1 {
		if mt == Add {
			return nil
		}
		op.perms.XCreate(
			ctx,
			utils.M{
				"role":  roleId,
				"based": basedRoleId,
			},
		)
		return nil
	}

	if mt == Add {
		return nil
	}

	q, vl = op.BindNamed("delete from %s where role=:rid and based=:brid", utils.M{"rid": roleId, "based": basedRoleId})
	op.inherits.XExecute(ctx, q, vl...)
	return nil
}

func (op *_RoleOp) getAllBasedRoles(ctx context.Context, role *_Role) {
	op.inherits.XQ1n(ctx, &role.Based, fmt.Sprintf(`select distinct based from %s where role=?`, op.inherits.TableName()), role.Id)
}

func (op *_RoleOp) List(ctx context.Context) (lst []*_Role) {
	for _, enum := range op.EnumOperator.List(ctx) {
		lst = append(lst, enum.(*_Role))
	}
	return
}
