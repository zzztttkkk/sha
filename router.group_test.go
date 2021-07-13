package sha

import (
	"fmt"
	"testing"
)

func TestNewRouteGroup(t *testing.T) {
	mux := NewMux(nil)

	handler := RequestCtxHandlerFunc(func(ctx *RequestCtx) {})

	groupA := NewRouteGroup("/a")
	groupAB := NewRouteGroup("/b")
	groupAB.HTTP(MethodGet, "/q", handler)
	groupAB.HTTP(MethodGet, "/w", handler)
	groupAB.HTTP(MethodGet, "/e", handler)

	groupAB.BindTo(groupA)

	groupA.HTTP(MethodGet, "/r", handler)

	groupA.BindTo(mux)

	fmt.Println(mux)
}

