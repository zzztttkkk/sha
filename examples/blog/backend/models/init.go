package models

import "github.com/zzztttkkk/snow"

var roleOpInitPriority = snow.NewPriority(10)
var permissionOpInitPriority = roleOpInitPriority.Copy()
var permissionCreatePriority = permissionOpInitPriority.Incr()
var rbacInitPriority = permissionCreatePriority.Incr()
