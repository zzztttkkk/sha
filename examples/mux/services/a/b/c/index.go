package c

import (
	"github.com/zzztttkkk/sha"
	"simple/h"
)

var Branch = sha.NewBranch()

type _H struct {
	Name string
}

func (h *_H) Handle(ctx *sha.RequestCtx) {
	_, _ = ctx.WriteString(h.Name)
}

func init() {
	Branch.Use(
		h.NewPrintMiddleware("a.b.c.m1"),
		h.NewPrintMiddleware("a.b.c.m2"),
		h.NewPrintMiddleware("a.b.c.m3"),
	)

	Branch.REST(
		"get",
		"/",
		h.NewPrintHandler("a.b.c.h"),
	)

	Branch.REST(
		"get",
		"/h",
		&_H{Name: "0.0"},
	)
}
