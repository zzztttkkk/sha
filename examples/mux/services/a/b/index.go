package b

import (
	"github.com/zzztttkkk/suna"
	"simple/services/a/b/c"
	"simple/services/h"
)

var Branch = suna.NewBranch()

func init() {
	Branch.AddBranch("/c", c.Branch)

	Branch.Use(
		h.NewPrintMiddleware("a.b.m1"),
		h.NewPrintMiddleware("a.b.m2"),
		h.NewPrintMiddleware("a.b.m3"),
	)

	Branch.AddHandler(
		"get",
		"/",
		h.NewPrintHandler("a.b.h"),
	)
}
