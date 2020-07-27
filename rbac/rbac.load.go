package rbac

import (
	"context"
	"sync"
)

var g sync.RWMutex

var permIdMap map[int64]*_Permission
var permNameMap map[string]*_Permission
var roleIdMap map[int64]*Role

var rolePermMap map[int64]map[int64]bool

func Load(ctx context.Context) {
	g.Lock()
	defer g.Unlock()

	permIdMap = map[int64]*_Permission{}
	permNameMap = map[string]*_Permission{}
	roleIdMap = map[int64]*Role{}
	rolePermMap = map[int64]map[int64]bool{}

	for _, enum := range _PermissionOperator.List(ctx) {
		p := enum.(*_Permission)
		if p.Id < 1 {
			continue
		}
		permIdMap[p.Id] = p
		permNameMap[p.Name] = p
	}

	for _, enum := range _RoleOperator.List(ctx) {
		r := enum.(*Role)
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

	_makeOneRole(role, map[int64]bool{}, pm, &err)

	if err {
		return nil
	}
	return pm
}

func _makeOneRole(role *Role, footprints map[int64]bool, permMap map[int64]bool, errPtr *bool) {
	_, ok := footprints[role.GetId()]
	if ok {
		*errPtr = true
		// todo log
		return
	}

	for _, pid := range role.Permissions {
		permMap[pid] = true
	}

	for _, basedId := range role.Based {
		rp := roleIdMap[basedId]
		if rp == nil {
			*errPtr = true
			// todo log
			return
		}
		_makeOneRole(rp, footprints, permMap, errPtr)
	}
}
