package router

import (
	"sync"

	fr "github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/zzztttkkk/snow/internal"
)

type _MiddlewareChainT struct {
	head    *[]fasthttp.RequestHandler
	cursor  int
	length  int
	handler fasthttp.RequestHandler
}

var handlersChainPool = &sync.Pool{
	New: func() interface{} {
		return &_MiddlewareChainT{
			head:    nil,
			cursor:  -1,
			length:  0,
			handler: nil,
		}
	},
}

func acquireHandlerChain() *_MiddlewareChainT {
	return handlersChainPool.Get().(*_MiddlewareChainT)
}

func releaseHandlerChain(hs *_MiddlewareChainT) {
	hs.head = nil
	hs.cursor = -1
	hs.length = 0
	hs.handler = nil
	handlersChainPool.Put(hs)
}

func Next(ctx *fasthttp.RequestCtx) {
	chain, ok := ctx.UserValue(internal.RCtxKeyHandlerChain).(*_MiddlewareChainT)
	if !ok || chain.head == nil {
		return
	}

	chain.cursor++
	if chain.cursor < chain.length {
		(*chain.head)[chain.cursor](ctx)
		return
	}

	if chain.cursor > chain.length {
		return
	}
	chain.handler(ctx)
}

type Router struct {
	*fr.Router
	parent     *Router
	middleware []fasthttp.RequestHandler
	prefix     string
	path       string
}

func New() *Router {
	return &Router{Router: fr.New(), path: "", prefix: ""}
}

func (router *Router) SubGroup(name string) *Router {
	path := "/" + name
	return &Router{
		Router: router.Router.Group(path),
		parent: router,
		prefix: path,
		path:   router.path + path,
	}
}

func (router *Router) Use(middleware ...fasthttp.RequestHandler) {
	router.middleware = append(router.middleware, middleware...)
}

func (router *Router) Handle(method string, path string, handler fasthttp.RequestHandler) {
	// get all super middleware middleware
	var allMiddleware [][]fasthttp.RequestHandler
	c := router
	for c.parent != nil {
		if c.parent.middleware != nil {
			allMiddleware = append(allMiddleware, c.parent.middleware)
		}
		c = c.parent
	}

	l, r := 0, len(allMiddleware)-1
	for l < r {
		allMiddleware[l], allMiddleware[r] = allMiddleware[r], allMiddleware[l]
		l++
		r--
	}

	if router.middleware != nil {
		allMiddleware = append(allMiddleware, router.middleware)
	}

	newHandler := handler
	if len(allMiddleware) > 0 {
		middleware := make([]fasthttp.RequestHandler, 0)
		for _, hs := range allMiddleware {
			middleware = append(middleware, hs...)
		}

		newHandler = func(ctx *fasthttp.RequestCtx) {
			chain := acquireHandlerChain()
			defer releaseHandlerChain(chain)

			chain.head = &middleware
			chain.length = len(middleware)
			chain.handler = handler

			ctx.SetUserValue(internal.RCtxKeyHandlerChain, chain)

			Next(ctx)
		}
	}
	router.Router.Handle(method, path, newHandler)
}

func (router *Router) GET(path string, handle fasthttp.RequestHandler) {
	router.Handle(fasthttp.MethodGet, path, handle)
}

func (router *Router) HEAD(path string, handle fasthttp.RequestHandler) {
	router.Handle(fasthttp.MethodHead, path, handle)
}

func (router *Router) OPTIONS(path string, handle fasthttp.RequestHandler) {
	router.Handle(fasthttp.MethodOptions, path, handle)
}

func (router *Router) POST(path string, handle fasthttp.RequestHandler) {
	router.Handle(fasthttp.MethodPost, path, handle)
}

func (router *Router) PUT(path string, handle fasthttp.RequestHandler) {
	router.Handle(fasthttp.MethodPut, path, handle)
}

func (router *Router) PATCH(path string, handle fasthttp.RequestHandler) {
	router.Handle(fasthttp.MethodPatch, path, handle)
}

func (router *Router) DELETE(path string, handle fasthttp.RequestHandler) {
	router.Handle(fasthttp.MethodDelete, path, handle)
}

func (router *Router) Path() string {
	if router.path == "" {
		return "/"
	}
	return router.path
}
