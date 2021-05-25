package sha

import (
	"embed"
	"net/http"
	"testing"
	"time"
)

func TestFs(t *testing.T) {
	server := Default()
	mux := NewMux(nil)

	mux.FileSystem(nil, "Get", "/sha/src/{filepath:*}", http.Dir("./"), true)

	server.Handler = mux
	server.ListenAndServe()
}

//go:embed *.go
var ef embed.FS

func TestNewEmbedFSHandler(t *testing.T) {
	ListenAndServe("", NewEmbedFSHandler(&ef, time.Time{}, func(ctx *RequestCtx) string { return ctx.Request.Path()[1:] }))
}
