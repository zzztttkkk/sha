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

	Branch.REST(
		"get",
		"/",
		h.NewPrintHandler("a.h"),
	)
}
