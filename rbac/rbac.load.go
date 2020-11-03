package rbac

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/suna/utils"
	"sync"
)

var g sync.RWMutex

var permIdMap map[int64]*Permission
var permNameMap map[string]*Permission
var roleIdMap map[int64]*Role
var errs []string
var rolePermMap map[int64]map[int64]bool
var rolePermCache = utils.NewLru(200)
var roleInheritMap map[int64]map[int64]bool

func Load(ctx context.Context) {
	g.Lock()
	defer g.Unlock()

	errs = make([]string, 0)
	permIdMap = map[int64]*Permission{}
	permNameMap = map[string]*Permission{}
	roleIdMap = map[int64]*Role{}
	rolePermMap = map[int64]map[int64]bool{}
	roleInheritMap = map[int64]map[int64]bool{}
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

func makeOneRole(role *Role) map[int64]bool {
	pm := map[int64]bool{}
	err := false
	fpm := map[int64]bool{}
	_makeOneRole(role, fpm, pm, &err)

	if err {
		return nil
	}
	for k, _ := range fpm {
		_m := roleInheritMap[role.Id]
		if _m == nil {
			_m = map[int64]bool{}
			roleInheritMap[k] = _m
		}
		_m[k] = true
	}
	return pm
}

func _makeOneRole(role *Role, footprints map[int64]bool, permMap map[int64]bool, errPtr *bool) {
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
