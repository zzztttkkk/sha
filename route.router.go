package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/utils"
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
	HTTPWithForm(method, path string, handler RequestHandler, form interface{})
	HTTPWithMiddleware(method, path string, handler RequestHandler, middleware ...Middleware)
	HTTPWithMiddlewareAndForm(method, path string, handler RequestHandler, form interface{}, middleware ...Middleware)

	WebSocket(path string, handler WebSocketHandlerFunc)

	FileSystem(fs http.FileSystem, method, path string, autoIndex bool, middleware ...Middleware)

	AddBranch(prefix string, router Router)

	Use(middleware ...Middleware)
}

const _filename = "filename"

func fileSystemHandler(fs http.FileSystem, path string, autoIndex bool, middleware ...Middleware) RequestHandler {
	if !strings.HasSuffix(path, "/filename:*") {
		panic(fmt.Errorf("sha.router: bad static path"))
	}
	return handlerWithMiddleware(
		RequestHandlerFunc(
			func(ctx *RequestCtx) {
				filename, _ := ctx.PathParam(_filename)
				serveFile(ctx, fs, pathlib.Clean(utils.S(filename)), autoIndex)
			},
		),
		middleware...,
	)
}
