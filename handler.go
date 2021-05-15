package sha

type RequestHandler interface {
	Handle(ctx *RequestCtx)
}

type RequestHandlerFunc func(ctx *RequestCtx)

func (fn RequestHandlerFunc) Handle(ctx *RequestCtx) { fn(ctx) }
