package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"net/http"
	"os"
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

	FilePath(fpath string, method, path string, autoIndex bool, middleware ...Middleware)
	File(fpath string, method, path string, middleware ...Middleware)

	AddBranch(prefix string, router Router)

	Use(middleware ...Middleware)
}

const _filename = "filename"

func makeFileSystemHandler(fpath string, path string, autoIndex bool, middleware ...Middleware) RequestHandler {
	if !strings.HasSuffix(path, "/filename:*") {
		panic(fmt.Errorf("sha.router: bad static path"))
	}
	return handlerWithMiddleware(
		RequestHandlerFunc(
			func(ctx *RequestCtx) {
				filename, _ := ctx.PathParam(_filename)
				serveFile(ctx, http.Dir(fpath), pathlib.Clean(utils.S(filename)), autoIndex)
			},
		),
		middleware...,
	)
}

func makeFileHandler(fpath string, middleware ...Middleware) RequestHandler {
	return handlerWithMiddleware(
		RequestHandlerFunc(func(ctx *RequestCtx) {
			f, err := os.Open(fpath)
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
			serveContent(ctx, d.Name(), d.ModTime(), d.Size(), f)
		}),
		middleware...,
	)
}
