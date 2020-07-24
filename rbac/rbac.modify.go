package rbac

import "context"

func NewPermission(ctx context.Context, name string) error {
	return backend.NewPermission(ctx, name)
}

func NewRole(ctx context.Context, name string) error {
	return backend.NewRole(ctx, name)
}

func RoleAddPermission(ctx context.Context, roleName, permName string) error {
	return backend.RoleAddPermission(ctx, roleName, permName)
}

func RoleDelPermission(ctx context.Context, roleName, permName string) error {
	return backend.RoleDelPermission(ctx, roleName, permName)
}

func SubjectAddRole(ctx context.Context, subject int64, roleName string) error {
	return backend.SubjectAddRole(ctx, subject, roleName)
}

func SubjectDelRole(ctx context.Context, subject int64, roleName string) error {
	return backend.SubjectDelRole(ctx, subject, roleName)
}

func SubjectAddPermission(ctx context.Context, subject int64, permissionName string) error {
	return backend.SubjectAddPermission(ctx, subject, permissionName)
}

func SubjectDelPermission(ctx context.Context, subject int64, permissionName string) error {
	return backend.SubjectDelPermission(ctx, subject, permissionName)
}
