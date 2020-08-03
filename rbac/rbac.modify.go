package rbac

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/utils"
)

// permission
func NewPermission(ctx context.Context, name, descp string) error {
	if _PermissionOperator.ExistsByName(ctx, name) {
		return output.HttpErrors[fasthttp.StatusBadRequest]
	}
	_PermissionOperator.Create(ctx, utils.M{"name": name, "descp": descp})
	return nil
}

func DelPermission(ctx context.Context, name string) error {
	if !_PermissionOperator.ExistsByName(ctx, name) {
		return output.HttpErrors[fasthttp.StatusNotFound]
	}
	_PermissionOperator.Delete(ctx, name)
	return nil
}

// role
func NewRole(ctx context.Context, name, descp string) error {
	if _RoleOperator.ExistsByName(ctx, name) {
		return output.HttpErrors[fasthttp.StatusBadRequest]
	}
	_RoleOperator.Create(ctx, utils.M{"name": name, "descp": descp})
	return nil
}

func DelRole(ctx context.Context, name string) error {
	if !_RoleOperator.ExistsByName(ctx, name) {
		return output.HttpErrors[fasthttp.StatusNotFound]
	}
	_RoleOperator.Delete(ctx, name)
	return nil
}

func RoleAddBased(ctx context.Context, a, based string) error {
	return _RoleOperator.changeInherits(ctx, a, based, Add)
}

func RoleDelBased(ctx context.Context, a, based string) error {
	return _RoleOperator.changeInherits(ctx, a, based, Del)
}

func RoleAddPerm(ctx context.Context, role, perm string) error {
	return _RoleOperator.changePerm(ctx, role, perm, Add)
}

func RoleDelPerm(ctx context.Context, role, perm string) error {
	return _RoleOperator.changePerm(ctx, role, perm, Del)
}

// user
func UserAddRole(ctx context.Context, uid int64, role string) error {
	return _UserOperator.changeRole(ctx, uid, role, Add)
}

func UserDelRole(ctx context.Context, uid int64, role string) error {
	return _UserOperator.changeRole(ctx, uid, role, Del)
}
