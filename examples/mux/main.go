package main

import (
	"fmt"
	"github.com/zzztttkkk/sha"
	"simple/h"
	"simple/services/a"
)

func main() {
	conf := sha.MuxOptions{
		DoTrailingSlashRedirect: true,
		AutoHandleOptions:       true,
	}

	mux := sha.NewMux(&conf)

	mux.Use(
		h.NewPrintMiddleware("g.m1"),
		h.NewPrintMiddleware("g.m2"),
		h.NewPrintMiddleware("g.m3"),
	)

	server := sha.Default()
	server.Handler = mux

	mux.AddGroup(a.Group)

	fmt.Println(mux)
	server.ListenAndServe()
}
