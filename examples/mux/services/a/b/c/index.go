package c

import (
	"github.com/zzztttkkk/suna"
	"simple/h"
)

var Branch = suna.NewBranch()

type _H struct {
	Name string
}

func (h *_H) Handle(ctx *suna.RequestCtx) {
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
