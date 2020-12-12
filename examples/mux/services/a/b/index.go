package b

import (
	"github.com/zzztttkkk/sha"
	"simple/h"
	"simple/services/a/b/c"
)

var Branch = sha.NewBranch()

func init() {
	Branch.AddBranch("/c", c.Branch)

	Branch.Use(
		h.NewPrintMiddleware("a.b.m1"),
		h.NewPrintMiddleware("a.b.m2"),
		h.NewPrintMiddleware("a.b.m3"),
	)

	Branch.HTTP(
		"get",
		"/",
		h.NewPrintHandler("a.b.h"),
	)
}
