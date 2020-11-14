package c

import (
	"github.com/zzztttkkk/suna"
	"simple/services/h"
)

var Branch = suna.NewBranch()

func init() {
	Branch.Use(
		h.NewPrintMiddleware("a.b.c.m1"),
		h.NewPrintMiddleware("a.b.c.m2"),
		h.NewPrintMiddleware("a.b.c.m3"),
	)

	Branch.AddHandler(
		"get",
		"/",
		h.NewPrintHandler("a.b.c.h"),
	)
}
