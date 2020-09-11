package rbac

import (
	"context"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
)

// permission
func NewPermission(ctx context.Context, name, descp string) error {
	if _PermissionOperator.ExistsByName(ctx, name) {
		return output.HttpErrors[fasthttp.StatusBadRequest]
	}
	_PermissionOperator.Create(ctx, name, descp)
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
	_RoleOperator.Create(ctx, name, descp)
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
	return _RoleOperator.changeInherits(ctx, a, based, _Add)
}

func RoleDelBased(ctx context.Context, a, based string) error {
	return _RoleOperator.changeInherits(ctx, a, based, _Del)
}

func RoleListAllBased(ctx context.Context, name string) ([]string, error) {
	_role, ok := _RoleOperator.GetByName(ctx, name)
	if !ok {
		return nil, output.HttpErrors[fasthttp.StatusNotFound]
	}

	role := Role{}
	role.Id = _role.GetId()
	_RoleOperator.getAllBasedRoles(ctx, &role)

	var lst []string
	for _, rid := range role.Based {
		_role, ok = _RoleOperator.GetById(ctx, rid)
		if !ok {
			return nil, output.HttpErrors[fasthttp.StatusInternalServerError]
		}
		lst = append(lst, _role.GetName())
	}
	return lst, nil
}

func RoleAddPerm(ctx context.Context, role, perm string) error {
	return _RoleOperator.changePerm(ctx, role, perm, _Add)
}

func RoleDelPerm(ctx context.Context, role, perm string) error {
	return _RoleOperator.changePerm(ctx, role, perm, _Del)
}

// user
func UserAddRole(ctx context.Context, uid int64, role string) error {
	return _UserOperator.changeRole(ctx, uid, role, _Add)
}

func UserDelRole(ctx context.Context, uid int64, role string) error {
	return _UserOperator.changeRole(ctx, uid, role, _Del)
}
