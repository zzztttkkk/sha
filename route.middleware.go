package suna

type Middleware interface {
	Process(ctx *RequestCtx, next func())
}

type MiddlewareFunc func(ctx *RequestCtx, next func())

func (fn MiddlewareFunc) Process(ctx *RequestCtx, next func()) {
	fn(ctx, next)
}

type _MiddlewareOrg struct {
	local      []Middleware
	allM       []Middleware
	parentMOrg *_MiddlewareOrg
	freezen    bool
}

func (org *_MiddlewareOrg) Use(middleware ...Middleware) {
	if org.freezen {
		panic("suna.router: router freezen")
	}
	org.local = append(org.local, middleware...)
}

func (org *_MiddlewareOrg) expand() {
	if len(org.allM) > 0 {
		return
	}
	if org.parentMOrg != nil {
		org.parentMOrg.expand()
		org.allM = append(org.allM, org.parentMOrg.allM...)
	}
	org.allM = append(org.allM, org.local...)
	org.freezen = true
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

func (org *_MiddlewareOrg) wrap(handler RequestHandler) RequestHandler {
	org.expand()
	if len(org.allM) < 1 {
		return handler
	}
	return handlerWithMiddleware(handler, org.allM...)
}
