package rbac

import (
	"context"
	"errors"
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/utils"
)

type CheckPolicy int

const (
	PolicyAll = CheckPolicy(iota)
	PolicyAny
)

var ErrUnknownRole = errors.New("suna.rbac: unknown role")

func getBitmap(roles utils.Int64Slice) *roaring64.Bitmap {
	if len(roles) < 1 {
		return nil
	}

	key := roles.Join(":")
	v, ok := rolePermCache.Get(key)
	if ok {
		return v.(*roaring64.Bitmap)
	}

	set := roaring64.New()

	for _, rid := range roles {
		m, ok := rolePermMap[rid]
		if !ok {
			panic(ErrUnknownRole)
		}
		for p := range m {
			set.Add(uint64(p))
		}
	}

	rolePermCache.Add(key, set)
	return set
}

//revive:disable:cyclomatic
func IsGranted(ctx context.Context, user auth.User, policy CheckPolicy, permissions ...string) bool {
	SubjectId := user.GetId()
	if SubjectId < 1 {
		return false
	}

	g.RLock()
	defer g.RUnlock()

	var Perms []uint64
	for _, name := range permissions {
		v, ok := permNameMap[name]
		if !ok {
			return false
		}
		if v.Id < 1 {
			return false
		}
		Perms = append(Perms, uint64(v.Id))
	}

	set := getBitmap(SubjectOperator.getRoles(ctx, SubjectId))
	if set == nil {
		return false
	}

	switch policy {
	case PolicyAll:
		for _, p := range Perms {
			if !set.Contains(p) {
				return false
			}
		}
		return true
	case PolicyAny:
		for _, p := range Perms {
			if set.Contains(p) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func SubjectHasRole(ctx context.Context, uid int64, roleName string) bool {
	return SubjectOperator.hasRole(ctx, uid, roleName)
}

func Permissions(ctx context.Context) []*Permission {
	return _PermissionOperator.List(ctx)
}

func Roles(ctx context.Context) []*Role {
	return _RoleOperator.List(ctx)
}
