package rbac

import (
	"context"
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/utils"
)

type CheckPolicy int

const (
	PolicyAll = CheckPolicy(iota)
	PolicyAny
)

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
			// todo log
			return nil
		}
		for p, _ := range m {
			set.Add(uint64(p))
		}
	}

	rolePermCache.Add(key, set)
	return set
}

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

	set := getBitmap(_UserOperator.getRoles(ctx, SubjectId))
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
