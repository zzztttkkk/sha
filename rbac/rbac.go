package rbac

import (
	"context"
	"errors"
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/golang/groupcache/lru"
	"github.com/zzztttkkk/sha/auth"
	shainternal "github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/rbac/dao"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var (
	g     sync.RWMutex
	perms map[string]int64
	roles map[string]int64
	rpm   map[int64]*roaring64.Bitmap
	cache lru.Cache
)

func clear() {
	perms = map[string]int64{}
	roles = map[string]int64{}
	cache.Clear()
	rpm = map[int64]*roaring64.Bitmap{}
}

func Load(ctx context.Context) {
	g.Lock()
	defer g.Unlock()
	clear()

	for _, p := range dao.Perms(ctx) {
		perms[p.Name] = p.ID
	}

	for _, r := range dao.Roles(ctx) {
		roles[r.Name] = r.ID

		bitmap := roaring64.New()
		for _, p := range dao.RolePermIDs(ctx, r.ID) {
			bitmap.Add(uint64(p))
		}
		rpm[r.ID] = bitmap
	}

	// merge perms
	for rid, bitmap := range rpm {
		for brid := range dao.GetAllBasedRoleIDs(ctx, rid) {
			t := rpm[brid]
			iter := t.Iterator()
			for iter.HasNext() {
				bitmap.Add(iter.Next())
			}
		}
	}
}

type _Policy int

const (
	all = _Policy(iota + 1)
	any
)

var ErrPermissionDenied = errors.New("sha.rbac: permission denied")
var ErrUnknownRole = errors.New("sha.rbac: unexpected role")
var ErrUnknownPermission = errors.New("sha.rbac: unknown permission")

func init() {
	shainternal.ErrorStatusByValue[ErrUnknownPermission] = http.StatusBadRequest
	shainternal.ErrorStatusByValue[ErrPermissionDenied] = http.StatusForbidden
	shainternal.ErrorStatusByValue[ErrUnknownRole] = http.StatusInternalServerError
}

func getBitmap(ctx context.Context, subject auth.Subject) *roaring64.Bitmap {
	var buf strings.Builder
	rs := dao.SubjectRoles(ctx, subject)
	for _, r := range rs {
		buf.WriteString(strconv.FormatInt(r, 10))
	}

	g.RLock()
	v, ok := cache.Get(buf.String())
	if ok {
		g.RUnlock()
		return v.(*roaring64.Bitmap)
	}
	g.RUnlock()

	g.Lock()
	defer g.Unlock()

	bitmap := roaring64.New()
	for _, r := range rs {
		t, ok := rpm[r]
		if !ok {
			panic(ErrUnknownRole)
		}
		bitmap.Or(t)
	}
	cache.Add(buf.String(), bitmap)
	return bitmap
}

func granted(ctx context.Context, policy _Policy, permissions ...string) error {
	subject, err := auth.Auth(ctx)
	if err != nil {
		return err
	}

	g.RLock()
	var ps []int64
	for _, pn := range permissions {
		pi, ok := perms[pn]
		if !ok {
			if policy == all {
				g.RUnlock()
				return ErrUnknownPermission
			}
			continue
		}
		ps = append(ps, pi)
	}
	g.RUnlock()

	if len(ps) < 1 {
		return ErrUnknownPermission
	}

	bitmap := getBitmap(ctx, subject)

	if policy == all {
		for _, pi := range ps {
			if !bitmap.Contains(uint64(pi)) {
				return ErrPermissionDenied
			}
		}
		return nil
	}

	for _, pi := range ps {
		if bitmap.Contains(uint64(pi)) {
			return nil
		}
	}
	return ErrPermissionDenied
}

func IsGrantedAll(ctx context.Context, permissions ...string) error {
	return granted(ctx, all, permissions...)
}

func IsGrantedAny(ctx context.Context, permissions ...string) error {
	return granted(ctx, any, permissions...)
}

func MustGrantedAll(ctx context.Context, permissions ...string) {
	if err := IsGrantedAll(ctx, permissions...); err != nil {
		panic(err)
	}
}

func MustGrantedAny(ctx context.Context, permissions ...string) {
	if err := IsGrantedAny(ctx, permissions...); err != nil {
		panic(err)
	}
}
