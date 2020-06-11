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
	Changed(ctx context.Context) bool // need reload
	LoadDone(ctx context.Context)
	GetAllPermissions(ctx context.Context) []Permission
	GetAllRoles(ctx context.Context) []Role
}

type Rbac interface {
	IsGranted(ctx context.Context, subject Subject, permission string) (bool, error)
	MustGranted(ctx context.Context, subject Subject, permission string)
}
