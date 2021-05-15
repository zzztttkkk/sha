package b

import (
	"github.com/zzztttkkk/sha"
	"simple/h"
	"simple/services/a/b/c"
)

var Group = sha.NewRouteGroup("/b")

func init() {
	Group.Use(
		h.NewPrintMiddleware("a.b.m1"),
		h.NewPrintMiddleware("a.b.m2"),
		h.NewPrintMiddleware("a.b.m3"),
	)
	Group.AddGroup(c.Group)

	Group.HTTP(
		"get",
		"/",
		h.NewPrintHandler("a.b.h"),
	)
}
