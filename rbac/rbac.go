package rbac

import "sync"

var backend Backend
var g sync.RWMutex

var MaxPrivateSubjectId int64

var permIdMap map[int64]Permission
var permNameMap map[string]Permission
var roleIdMap map[int64]Role

var rolePermMap map[int64]map[int64]bool

func SetBackend(v Backend) {
	g.Lock()
	defer g.Unlock()

	backend = v
}
