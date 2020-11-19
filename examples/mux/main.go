package main

import (
	"context"
	"github.com/zzztttkkk/suna"
	"simple/h"
	"simple/services/a"
)

func main() {
	server := suna.Server{
		Host:    "127.0.0.1",
		Port:    8080,
		BaseCtx: context.Background(),
	}

	mux := suna.NewMux("", nil)
	mux.AutoRedirect = true
	mux.AutoOptions = true

	mux.Use(
		h.NewPrintMiddleware("g.m1"),
		h.NewPrintMiddleware("g.m2"),
		h.NewPrintMiddleware("g.m3"),
	)

	mux.AddBranch("/a", a.Branch)

	server.Handler = mux
	server.ListenAndServe()
}
