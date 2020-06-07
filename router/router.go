package router

import (
	fr "github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"sync"
)

type _HandlerChainT struct {
	head    *[]fasthttp.RequestHandler
	cursor  int
	length  int
	handler fasthttp.RequestHandler
}

const handlersKey = "/uhs"

var handlersChainPool = &sync.Pool{
	New: func() interface{} {
		return &_HandlerChainT{
			head:    nil,
			cursor:  -1,
			length:  0,
			handler: nil,
		}
	},
}

func putHandlers(hs *_HandlerChainT) {
	hs.head = nil
	hs.cursor = -1
	hs.length = 0
	hs.handler = nil
	handlersChainPool.Put(hs)
}

func Next(ctx *fasthttp.RequestCtx) {
	handlers, ok := ctx.UserValue(handlersKey).(*_HandlerChainT)
	if !ok {
		return
	}

	handlers.cursor++
	if handlers.cursor < handlers.length {
		(*handlers.head)[handlers.cursor](ctx)
		return
	}

	if handlers.cursor == handlers.length {
		handlers.handler(ctx)
	}
}

type Router struct {
	*fr.Router
	parent   *Router
	handlers []fasthttp.RequestHandler
	prefix   string
	path     string
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
	router.handlers = append(router.handlers, middleware...)
}

func (router *Router) Handle(method string, path string, handler fasthttp.RequestHandler) {
	// get all super middleware handlers
	allHandlerSlice := make([][]fasthttp.RequestHandler, 0)
	c := router
	for c.parent != nil {
		if c.parent.handlers != nil {
			allHandlerSlice = append(allHandlerSlice, c.parent.handlers)
		}
		c = c.parent
	}

	l, r := 0, len(allHandlerSlice)-1
	for l < r {
		allHandlerSlice[l], allHandlerSlice[r] = allHandlerSlice[r], allHandlerSlice[l]
		l++
		r--
	}

	if router.handlers != nil {
		allHandlerSlice = append(allHandlerSlice, router.handlers)
	}

	newHandler := handler
	if len(allHandlerSlice) > 0 {
		middleware := make([]fasthttp.RequestHandler, 0)
		for _, hs := range allHandlerSlice {
			middleware = append(middleware, hs...)
		}

		newHandler = func(ctx *fasthttp.RequestCtx) {
			_hs := handlersChainPool.Get().(*_HandlerChainT)
			defer putHandlers(_hs)

			_hs.head = &middleware
			_hs.length = len(middleware)
			_hs.handler = handler

			ctx.SetUserValue(handlersKey, _hs)

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
