package main

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/sha"
	"os/signal"
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

	ctx, cancelFunc := signal.NotifyContext(context.Background())
	defer cancelFunc()
	server := sha.DefaultWithContext(ctx)
	server.Options.Pid = "./sha.mux.pid"
	server.Handler = mux

	mux.AddGroup(a.Group)

	fmt.Println(mux)
	server.ListenAndServe()
}
