package interfaces

import "github.com/zzztttkkk/suna/rbac"

type User interface {
	rbac.Subject
}
