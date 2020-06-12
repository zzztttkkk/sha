package mware

import "github.com/zzztttkkk/snow/rbac"

type User interface {
	rbac.Subject
}
