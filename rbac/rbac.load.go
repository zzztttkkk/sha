package rbac

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/suna/cache"
	"sync"
	"sync/atomic"
	"time"
)

var g sync.RWMutex

var permIdMap map[int64]*_Permission
var permNameMap map[string]*_Permission
var roleIdMap map[int64]*_Role
var errs []string
var rolePermMap map[int64]map[int64]bool
var rolePermCache = cache.NewLru(200)

func Load(ctx context.Context) {
	g.Lock()
	defer g.Unlock()

	errs = make([]string, 0)
	permIdMap = map[int64]*_Permission{}
	permNameMap = map[string]*_Permission{}
	roleIdMap = map[int64]*_Role{}
	rolePermMap = map[int64]map[int64]bool{}
	rolePermCache.Clear()

	for _, p := range _PermissionOperator.List(ctx) {
		if p.Id < 1 {
			continue
		}
		permIdMap[p.Id] = p
		permNameMap[p.Name] = p
	}

	for _, r := range _RoleOperator.List(ctx) {
		if r.Id < 1 {
			continue
		}
		roleIdMap[r.Id] = r
	}

	buildRolePermMap()
}

func buildRolePermMap() {
	for rid, role := range roleIdMap {
		rolePermMap[rid] = makeOneRole(role)
	}
}

func makeOneRole(role *_Role) map[int64]bool {
	pm := map[int64]bool{}
	err := false

	_makeOneRole(role, map[int64]bool{}, pm, &err)

	if err {
		return nil
	}
	return pm
}

func _makeOneRole(role *_Role, footprints map[int64]bool, permMap map[int64]bool, errPtr *bool) {
	_, ok := footprints[role.GetId()]
	if ok {
		*errPtr = true
		errs = append(errs, fmt.Sprintf("circular reference: role `%s`", role.Name))
		return
	}

	for _, pid := range role.Permissions {
		permMap[pid] = true
	}

	for _, basedId := range role.Based {
		rp := roleIdMap[basedId]
		if rp == nil {
			*errPtr = true
			errs = append(errs, fmt.Sprintf("not exists: role `%d`", basedId))
			return
		}
		_makeOneRole(rp, footprints, permMap, errPtr)
	}
}

var changeCount int64
var loading int32

func reload() {
	if v := recover(); v != nil {
		panic(v)
	}

	if atomic.LoadInt32(&loading) > 0 {
		atomic.AddInt64(&changeCount, 1)
		return
	}
	atomic.StoreInt32(&loading, 1)
	defer atomic.StoreInt32(&loading, 0)

	atomic.StoreInt64(&changeCount, 0)
	Load(context.Background())

	if atomic.LoadInt64(&changeCount) > 0 {
		atomic.StoreInt64(&changeCount, 0)
		Load(context.Background())
		time.Sleep(time.Millisecond * 500)
	}
}
