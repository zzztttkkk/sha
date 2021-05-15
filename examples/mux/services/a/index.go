package a

import (
	"github.com/zzztttkkk/sha"
	"simple/h"
	"simple/services/a/b"
)

var Group = sha.NewRouteGroup("/a")

func init() {
	Group.Use(
		h.NewPrintMiddleware("a.m1"),
		h.NewPrintMiddleware("a.m2"),
		h.NewPrintMiddleware("a.m3"),
	)
	Group.AddGroup(b.Group)

	Group.HTTPWithOptions(
		&sha.HandlerOptions{
			Middlewares: []sha.Middleware{
				h.NewPrintMiddleware("a.s.m1"),
				h.NewPrintMiddleware("a.s.m2"),
			},
		},
		"get",
		"",
		h.NewPrintHandler("a.h"),
	)
}
