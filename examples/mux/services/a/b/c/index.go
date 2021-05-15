package c

import (
	"github.com/zzztttkkk/sha"
	"simple/h"
)

var Group = sha.NewRouteGroup("/c")

func init() {
	Group.Use(
		h.NewPrintMiddleware("a.b.c.m1"),
		h.NewPrintMiddleware("a.b.c.m2"),
		h.NewPrintMiddleware("a.b.c.m3"),
	)

	Group.HTTP(
		"get",
		"/",
		h.NewPrintHandler("a.b.c.h"),
	)
}
