package sha

import (
	"context"
	"io"
	"net/http"

	"github.com/zzztttkkk/sha/utils"
)

func (ctx *RequestCtx) toStdRequest() *http.Request {
	req := &ctx.Request
	var body io.Reader
	if req.body != nil {
		body = req.body
	}

	var _ctx context.Context = ctx
	req.URL.Params.EachItem(func(item *utils.KvItem) bool {
		//lint:ignore SA1029 nothing
		_ctx = context.WithValue(_ctx, string(item.Key), string(item.Val))
		return true
	})
	std, _ := http.NewRequestWithContext(_ctx, string(req.Method()), string(req.RawPath()), body)
	req.Header().EachItem(func(item *utils.KvItem) bool {
		std.Header.Add(string(item.Key), string(item.Val))
		return true
	})
	return std
}

type _ResponseWriter struct {
	res    *Response
	header http.Header
}

func (rw *_ResponseWriter) Header() http.Header {
	if rw.header == nil {
		rw.header = map[string][]string{}
	}
	return rw.header
}

func (rw *_ResponseWriter) Write(bytes []byte) (int, error) { return rw.res.Write(bytes) }

func (rw *_ResponseWriter) WriteHeader(statusCode int) { rw.res.SetStatusCode(statusCode) }

var _ http.ResponseWriter = (*_ResponseWriter)(nil)

func (ctx *RequestCtx) toStdResponse() *_ResponseWriter {
	return &_ResponseWriter{&ctx.Response, nil}
}

func WrapStdHandler(handler http.Handler) RequestHandler {
	return RequestHandlerFunc(func(ctx *RequestCtx) {
		res := ctx.toStdResponse()
		defer func() {
			if res.header == nil {
				return
			}
			ctx.Response.Header().LoadMap(utils.MultiValueMap(res.header))
		}()
		handler.ServeHTTP(res, ctx.toStdRequest())
	})
}

func WrapStdHandlerFunc(fn func(w http.ResponseWriter, r *http.Request)) RequestHandler {
	return WrapStdHandler(http.HandlerFunc(fn))
}
