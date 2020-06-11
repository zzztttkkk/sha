package rbac

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/snow/utils"
	"golang.org/x/sync/singleflight"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type _InstanceT struct {
	sg      singleflight.Group
	rwm     sync.RWMutex
	backend Backend

	permissionIdMap   map[uint32]Permission
	permissionNameMap map[string]Permission

	roleNameMap map[string]Role
	roleIdMap   map[uint32]Role

	roleBitmapMap   map[uint32]*utils.Bitmap
	roleBitmapCache map[string]*utils.Bitmap

	prev *_InstanceT
}

func (ins *_InstanceT) doLoad(ctx context.Context) error {
	ins.rwm.Lock()

	defer func() {
		ins.rwm.Unlock()
		ins.backend.LoadDone(ctx)
	}()

	ins.permissionIdMap = map[uint32]Permission{}
	ins.permissionNameMap = map[string]Permission{}

	ins.roleNameMap = map[string]Role{}
	ins.roleIdMap = map[uint32]Role{}

	ins.roleBitmapCache = map[string]*utils.Bitmap{}
	ins.roleBitmapMap = map[uint32]*utils.Bitmap{}

	var maxPermissionId uint32 = 0
	for _, permission := range ins.backend.GetAllPermissions(ctx) {
		ins.permissionIdMap[permission.GetId()] = permission
		ins.permissionNameMap[permission.GetName()] = permission

		if permission.GetId() > maxPermissionId {
			maxPermissionId = permission.GetId()
		}
	}

	for _, role := range ins.backend.GetAllRoles(ctx) {
		ins.roleNameMap[role.GetName()] = role
		ins.roleIdMap[role.GetId()] = role
	}

	for _, r := range ins.roleNameMap {
		bitmap := utils.NewBitmap(maxPermissionId + 1)
		err := ins.makePermissionBitmap(r, bitmap, []string{})
		if err != nil {
			return err
		}

		ins.roleBitmapMap[r.GetId()] = bitmap
		ins.roleBitmapCache[strconv.FormatUint(uint64(r.GetId()), 10)] = bitmap
	}

	return nil
}

func (ins *_InstanceT) getPermissionByName(name string) Permission {
	return ins.permissionNameMap[name]
}

func (ins *_InstanceT) getPermissionById(id uint32) Permission {
	return ins.permissionIdMap[id]
}

func (ins *_InstanceT) getRoleByName(name string) Role {
	return ins.roleNameMap[name]
}

func (ins *_InstanceT) getRoleById(id uint32) Role {
	return ins.roleIdMap[id]
}

func (ins *_InstanceT) load(ctx context.Context) error {
	_, err, _ := ins.sg.Do(
		"DOLOAD",
		func() (interface{}, error) {
			err := ins.doLoad(ctx)
			return nil, err
		},
	)
	return err
}

func (ins *_InstanceT) makePermissionBitmap(role Role, bitmap *utils.Bitmap, path []string) error {
	for _, name := range path {
		if name == role.GetName() {
			return fmt.Errorf("snow.rbac: circly reference, Role<`%s`>", role.GetName())
		}
	}

	parentId := role.GetParentId()
	if parentId > 0 {
		parent := ins.getRoleById(parentId)
		if parent == nil {
			return fmt.Errorf("snow.rbac: nil role, %d", parentId)
		}
		err := ins.makePermissionBitmap(parent, bitmap, append(path, role.GetName()))
		if err != nil {
			return err
		}
	}

	for _, permission := range strings.Split(role.GetPermissionIds(), ",") {
		_pid, err := strconv.ParseUint(permission, 10, 32)
		if err != nil {
			return err
		}
		pid := uint32(_pid)
		if p := ins.getPermissionById(pid); p == nil {
			return fmt.Errorf("snow.rbac: nil permission, %d", pid)
		}
		bitmap.Add(pid)
	}
	return nil
}

func (ins *_InstanceT) mergeRoles(roles _RolesT) *utils.Bitmap {
	sort.Sort(roles)

	key := ""
	last := len(roles) - 1
	for ind, role := range roles {
		key += strconv.FormatUint(uint64(role.GetId()), 10)
		if ind < last {
			key += "&"
		}
	}

	ins.rwm.RLock()
	bitmap, ok := ins.roleBitmapCache[key]
	if ok {
		ins.rwm.RUnlock()
		return bitmap
	}

	var result *utils.Bitmap
	for _, role := range roles {
		bitmap = ins.roleBitmapMap[role.GetId()]
		if result == nil {
			result = bitmap
		} else {
			result = result.Or(bitmap)
		}
	}

	ins.rwm.RUnlock()
	ins.rwm.Lock()
	ins.roleBitmapCache[key] = result
	ins.rwm.Unlock()

	return result
}

type _RolesT []Role

func (rs _RolesT) Len() int { return len(rs) }

func (rs _RolesT) Less(i, j int) bool { return rs[i].GetName() < rs[j].GetName() }

func (rs _RolesT) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }

func (ins *_InstanceT) IsGranted(ctx context.Context, subject Subject, permission string) (bool, error) {
	if ins.backend.Changed(ctx) {
		err := ins.load(ctx)
		if err != nil {
			panic(err)
		}
	}

	ins.rwm.RLock()

	p := ins.getPermissionByName(permission)
	if p == nil {
		ins.rwm.RUnlock()
		return false, fmt.Errorf("snow.rbac: nil permission, `%s`", permission)
	}
	var roles _RolesT
	for _, _rid := range strings.Split(subject.GetRoleIds(), ",") {
		_rid, err := strconv.ParseUint(_rid, 10, 32)
		if err != nil {
			ins.rwm.RUnlock()
			return false, err
		}
		role := ins.getRoleById(uint32(_rid))
		if role == nil {
			ins.rwm.RUnlock()
			return false, fmt.Errorf("snow.rbac: nil role, %d", _rid)
		}
		roles = append(roles, role)
	}
	ins.rwm.RUnlock()
	return ins.mergeRoles(roles).Has(p.GetId()), nil
}

func (ins *_InstanceT) MustGranted(ctx context.Context, subject Subject, permission string) {
	ok, err := ins.IsGranted(ctx, subject, permission)
	if ok {
		return
	}

	if err != nil {
		panic(err)
	}

	panic(&Error{SubjectId: subject.GetId(), PermissionName: permission})
}

func Default(backend Backend) Rbac {
	ins := &_InstanceT{backend: backend}
	err := ins.doLoad(context.Background())
	if err != nil {
		panic(err)
	}
	return ins
}
