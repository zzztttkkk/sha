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

func (org *_MiddlewareOrg) wrap(handler RequestHandler) RequestHandler {
	org.expand()
	if len(org.allM) < 1 {
		return handler
	}
	return RequestHandlerFunc(
		func(ctx *RequestCtx) {
			var cursor = -1
			var next func()
			var size = len(org.allM)
			next = func() {
				cursor++
				if cursor < size {
					org.allM[cursor].Process(ctx, next)
				} else {
					handler.Handle(ctx)
				}
			}
			next()
		},
	)
}
