package main

import (
	"github.com/zzztttkkk/sha"
	"simple/h"
	"simple/services/a"
)

func main() {
	conf := sha.MuxOption{
		AutoSlashRedirect: true,
		AutoOptions:       true,
	}

	mux := sha.NewMux(&conf, nil)

	mux.Use(
		h.NewPrintMiddleware("g.m1"),
		h.NewPrintMiddleware("g.m2"),
		h.NewPrintMiddleware("g.m3"),
	)

	server := sha.Default(mux)

	mux.AddBranch("/a", a.Branch)

	mux.Print()
	server.ListenAndServe()
}
