package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/internal"
	"net/http"
	pathlib "path"
	"strings"
)

type Documenter interface {
	Document() string
}

// request Handler with document
type DocedRequestHandler interface {
	Documenter
	RequestHandler
}

type Router interface {
	HTTP(method, path string, handler RequestHandler)
	WebSocket(path string, handler WebSocketHandlerFunc)
	HTTPWithForm(method, path string, handler RequestHandler, form interface{})
	FileSystem(fs http.FileSystem, method, path string, autoIndex bool, middleware ...Middleware)
	AddBranch(prefix string, router Router)
	Use(middleware ...Middleware)
}

func fileSystemHandler(fs http.FileSystem, path string, autoIndex bool, middleware ...Middleware) RequestHandler {
	if !strings.HasSuffix(path, "/filename:*") {
		panic(fmt.Errorf("sha.router: bad static path"))
	}
	return handlerWithMiddleware(
		RequestHandlerFunc(
			func(ctx *RequestCtx) {
				filename, _ := ctx.PathParam(internal.B("filename"))
				serveFile(ctx, fs, pathlib.Clean(internal.S(filename)), autoIndex)
			},
		),
		middleware...,
	)
}
