package rbac

import "context"

type Permission interface {
	GetId() uint32
	GetName() string
}

type Role interface {
	GetId() uint32
	GetName() string
	GetParentId() uint32
	GetPermissionIds() string
}

type Subject interface {
	GetId() int64
	GetRoleIds() string
}

type Backend interface {
	GetAllPermissions(ctx context.Context) []Permission
	GetAllRoles(ctx context.Context) []Role
}

type Rbac interface {
	Reload(ctx context.Context)
	IsValid(ctx context.Context) (bool, error)
	IsGranted(ctx context.Context, subject Subject, permission string) (bool, error)
	MustGranted(ctx context.Context, subject Subject, permission string)
}
