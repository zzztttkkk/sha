package sha

import (
	"net/http/pprof"
	"testing"
)

func TestWrapStdHandler(t *testing.T) {
	HandleFunc(MethodGet, "/debug/pprof/", WrapStdHandlerFunc(pprof.Index))
	HandleFunc(MethodGet, "/debug/pprof/cmdline", WrapStdHandlerFunc(pprof.Cmdline))
	HandleFunc(MethodGet, "/debug/pprof/profile", WrapStdHandlerFunc(pprof.Profile))
	HandleFunc(MethodGet, "/debug/pprof/symbol", WrapStdHandlerFunc(pprof.Symbol))
	HandleFunc(MethodGet, "/debug/pprof/trace", WrapStdHandlerFunc(pprof.Trace))

	ListenAndServe("", nil)
}
