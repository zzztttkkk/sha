package suna

type Middleware interface {
	Process(ctx *RequestCtx, next func())
}

type MiddlewareFunc func(ctx *RequestCtx, next func())

func (fn MiddlewareFunc) Process(ctx *RequestCtx, next func()) {
	fn(ctx, next)
}

type _MiddlewareNode struct {
	local        []Middleware
	allM         []Middleware
	parentMwNode *_MiddlewareNode
	frozen       bool
}

func (org *_MiddlewareNode) Use(middleware ...Middleware) {
	if org.frozen {
		panic("suna.router: has been frozen")
	}
	org.local = append(org.local, middleware...)
}

func (org *_MiddlewareNode) expand() {
	if len(org.allM) > 0 {
		return
	}
	if org.parentMwNode != nil {
		org.parentMwNode.expand()
		org.allM = append(org.allM, org.parentMwNode.allM...)
	}
	org.allM = append(org.allM, org.local...)
	org.frozen = true
}

type _MiddlewareWrapper struct {
	middleware []Middleware
	raw        RequestHandler
}

func (w *_MiddlewareWrapper) Handle(ctx *RequestCtx) {
	var next func()
	cursor := -1
	next = func() {
		cursor++
		if cursor < len(w.middleware) {
			w.middleware[cursor].Process(ctx, next)
		} else {
			w.raw.Handle(ctx)
		}
	}
	next()
}

func handlerWithMiddleware(handler RequestHandler, middleware ...Middleware) RequestHandler {
	if len(middleware) < 1 {
		return handler
	}
	return &_MiddlewareWrapper{
		middleware: middleware,
		raw:        handler,
	}
}

func (org *_MiddlewareNode) wrap(handler RequestHandler) RequestHandler {
	org.expand()
	if len(org.allM) < 1 {
		return handler
	}
	return handlerWithMiddleware(handler, org.allM...)
}
