package a

import (
	"github.com/zzztttkkk/sha"
	"simple/h"
	"simple/services/a/b"
)

var Branch = sha.NewBranch()

func init() {
	Branch.AddBranch("/b", b.Branch)

	Branch.Use(
		h.NewPrintMiddleware("a.m1"),
		h.NewPrintMiddleware("a.m2"),
		h.NewPrintMiddleware("a.m3"),
	)

	Branch.HTTPWithMiddleware(
		[]sha.Middleware{
			h.NewPrintMiddleware("a.s.m1"),
			h.NewPrintMiddleware("a.s.m2"),
		},
		"get",
		"/",
		h.NewPrintHandler("a.h"),
	)
}
