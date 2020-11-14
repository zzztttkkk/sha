package suna

type Router interface {
	AddHandler(method, path string, handler RequestHandler)
	AddBranch(prefix string, router Router)
	Use(middleware ...Middleware)
}

type Documenter interface {
	Document() string
}

type DocedRequestHandler interface {
	Documenter
	RequestHandler
}

type _Dh struct {
	Documenter
	h func(ctx *RequestCtx)
}

func (dn *_Dh) Handle(ctx *RequestCtx) {
	dn.h(ctx)
}

func NewDocedRequestHandler(fn func(ctx *RequestCtx), doc Documenter) DocedRequestHandler {
	return &_Dh{Documenter: doc, h: fn}
}
