package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/validator"
	"net/http"
)

type Middleware interface {
	Process(ctx *RequestCtx, next func())
}

type MiddlewareFunc func(ctx *RequestCtx, next func())

func (f MiddlewareFunc) Process(ctx *RequestCtx, next func()) {
	f(ctx, next)
}

type HandlerOptions struct {
	Middlewares []Middleware
	Document    validator.Document
}

type Router interface {
	HTTPWithOptions(opt *HandlerOptions, method, path string, handler RequestHandler)
	HTTP(method, path string, handler RequestHandler)
	HTTPWithForm(method, path string, handler RequestHandler, form interface{})
	Websocket(path string, handlerFunc WebsocketHandlerFunc, opt *HandlerOptions)
	FileSystem(opt *HandlerOptions, method, path string, fs http.FileSystem, autoIndex bool)
	File(opt *HandlerOptions, method, path, filepath string)

	Use(middlewares ...Middleware)
	Frozen()

	NewGroup(prefix string) Router
	AddGroup(group *MuxGroup)
}

func middlewaresWrap(middlewares []Middleware, h RequestHandler) RequestHandler {
	return RequestHandlerFunc(func(ctx *RequestCtx) {
		cursor := -1

		var next func()

		next = func() {
			if ctx.err != nil {
				return
			}

			cursor++
			if cursor < len(middlewares) {
				middlewares[cursor].Process(ctx, next)
			} else {
				h.Handle(ctx)
			}
		}
		next()
	})
}

type _MiddlewareNode struct {
	p      *_MiddlewareNode
	local  []Middleware
	frozen bool
}

func (m *_MiddlewareNode) Use(middlewares ...Middleware) {
	if m.frozen {
		panic(fmt.Errorf("sha.router: Router's middleware should registerd first"))
	}
	m.local = append(m.local, middlewares...)
}
