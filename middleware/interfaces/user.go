package interfaces

import "github.com/zzztttkkk/snow/rbac"

type User interface {
	rbac.Subject
}
