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

	FilePath(fs http.FileSystem, method, path string, autoIndex bool, middleware ...Middleware)
	File(fs http.FileSystem, filename, method, path string, middleware ...Middleware)

	AddBranch(prefix string, router Router)

	Use(middleware ...Middleware)
}

const _filename = "filename"

func makeFileSystemHandler(fs http.FileSystem, path string, autoIndex bool, middleware ...Middleware) RequestHandler {
	if !strings.HasSuffix(path, "/filename:*") {
		panic(fmt.Errorf("sha.router: bad static path"))
	}
	return handlerWithMiddleware(
		RequestHandlerFunc(
			func(ctx *RequestCtx) {
				filename, _ := ctx.PathParam(_filename)
				ServeFileSystem(ctx, fs, pathlib.Clean(utils.S(filename)), autoIndex)
			},
		),
		middleware...,
	)
}

func makeFileHandler(fs http.FileSystem, filename string, middleware ...Middleware) RequestHandler {
	return handlerWithMiddleware(
		RequestHandlerFunc(func(ctx *RequestCtx) {
			f, err := fs.Open(filename)
			if err != nil {
				ctx.SetStatus(toHTTPError(err))
				return
			}
			defer f.Close()

			d, err := f.Stat()
			if err != nil {
				ctx.SetStatus(toHTTPError(err))
				return
			}
			ServeFileContent(ctx, d.Name(), d.ModTime(), d.Size(), f)
		}),
		middleware...,
	)
}
