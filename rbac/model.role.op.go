package rbac

import (
	"context"
	"fmt"

	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/sqls"
	"github.com/zzztttkkk/suna/utils"
)

type roleOpT struct {
	sqls.EnumOperator
	perms    *sqls.Operator
	inherits *sqls.Operator
}

var _RoleOperator = &roleOpT{
	perms:    &sqls.Operator{},
	inherits: &sqls.Operator{},
}

func init() {
	dig.Provide(
		func(_ _DigPermissionTableInited) _DigRoleTableInited {
			_RoleOperator.perms.Init(roleWithPermT{})
			_RoleOperator.inherits.Init(roleInheritanceT{})

			_RoleOperator.EnumOperator.Init(
				Role{},
				func() sqls.EnumItem { return &Role{} },
				func(ctx context.Context, i interface{}) error {
					role := i.(*Role)
					_RoleOperator.getAllPerms(ctx, role)
					_RoleOperator.getAllBasedRoles(ctx, role)
					return nil
				},
			)
			return _DigRoleTableInited(0)
		},
	)
}

func (op *roleOpT) changePerm(ctx context.Context, roleName, permName string, mt modifyType) error {
	OP := op.perms

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

	cond := sqls.STR("role=? and perm=?", roleId, permId)

	var _id int64
	OP.ExecSelect(ctx, &_id, sqls.Select("id").Where(cond))
	if _id < 1 {
		if mt == _Add {
			return nil
		}

		OP.ExecInsert(
			ctx,
			sqls.Insert("perm, role").Values(permId, roleId),
		)
		return nil
	}

	if mt == _Add {
		return nil
	}
	OP.ExecDelete(ctx, sqls.Delete().Where(cond).Limit(1))
	return nil
}

func (op *roleOpT) getAllPerms(ctx context.Context, role *Role) {
	OP := op.perms
	OP.ExecSelect(
		ctx,
		&role.Permissions,
		sqls.Select("perm").Distinct().Where("role=?", role.Id),
	)
}

func (op *roleOpT) changeInherits(ctx context.Context, roleName, basedRoleName string, mt modifyType) error {
	OP := op.inherits

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

	cond := sqls.STR("role=? and based=?", roleId, basedRoleId)

	var _id int64
	OP.ExecSelect(ctx, &_id, sqls.Select("based").Where(cond))
	if _id < 1 {
		if mt == _Add {
			return nil
		}
		OP.ExecInsert(
			ctx,
			sqls.Insert("role,based").Values(roleId, basedRoleId),
		)
		return nil
	}

	if mt == _Add {
		return nil
	}

	OP.ExecDelete(ctx, sqls.Delete().Where(cond).Limit(1))
	return nil
}

func (op *roleOpT) getAllBasedRoles(ctx context.Context, role *Role) {
	OP := op.inherits
	OP.ExecSelect(
		ctx,
		&role.Based,
		sqls.Select("based").Distinct().Where("role=?", role.Id),
	)
}

func (op *roleOpT) List(ctx context.Context) (lst []*Role) {
	for _, enum := range op.EnumOperator.List(ctx) {
		lst = append(lst, enum.(*Role))
	}
	return
}
