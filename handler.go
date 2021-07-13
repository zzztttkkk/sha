package sha

type RequestCtxHandler interface {
	Handle(ctx *RequestCtx)
}

type RequestCtxHandlerFunc func(ctx *RequestCtx)

func (fn RequestCtxHandlerFunc) Handle(ctx *RequestCtx) { fn(ctx) }
