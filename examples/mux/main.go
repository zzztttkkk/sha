package main

import (
	"github.com/zzztttkkk/suna"
	"simple/h"
	"simple/services/a"
)

func main() {
	mux := suna.NewMux("", nil)
	mux.AutoSlashRedirect = true
	mux.AutoOptions = true

	mux.Use(
		h.NewPrintMiddleware("g.m1"),
		h.NewPrintMiddleware("g.m2"),
		h.NewPrintMiddleware("g.m3"),
	)

	server := suna.Default(mux)

	mux.AddBranch("/a", a.Branch)

	mux.Print(true, true)
	server.ListenAndServe()
}
