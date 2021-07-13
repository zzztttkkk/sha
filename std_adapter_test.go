package sha

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"testing"
	"unsafe"
)

func TestWrapStdHandler(t *testing.T) {
	HandleFunc(MethodGet, "/debug/pprof/", WrapStdHandlerFunc(pprof.Index))
	HandleFunc(MethodGet, "/debug/pprof/cmdline", WrapStdHandlerFunc(pprof.Cmdline))
	HandleFunc(MethodGet, "/debug/pprof/profile", WrapStdHandlerFunc(pprof.Profile))
	HandleFunc(MethodGet, "/debug/pprof/symbol", WrapStdHandlerFunc(pprof.Symbol))
	HandleFunc(MethodGet, "/debug/pprof/trace", WrapStdHandlerFunc(pprof.Trace))

	ListenAndServe("", nil)
}

func TestStdServer(t *testing.T) {
	server := &http.Server{Addr: "127.0.0.1:8080"}
	server.Handler = ToStdHandler(RequestCtxHandlerFunc(func(ctx *RequestCtx) {
		_ = ctx.WriteString("Hello world!")
		fmt.Println(ctx.Request.Path(), uintptr(unsafe.Pointer(ctx)))
	}))
	_ = server.ListenAndServe()
}
