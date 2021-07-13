package sha

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
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

func WrapStdHandler(handler http.Handler) RequestCtxHandler {
	return RequestCtxHandlerFunc(func(ctx *RequestCtx) {
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

func WrapStdHandlerFunc(fn func(w http.ResponseWriter, r *http.Request)) RequestCtxHandler {
	return WrapStdHandler(http.HandlerFunc(fn))
}

func ToStdHandler(handler RequestCtxHandler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		rctx := AcquireRequestCtx(ctx)
		defer ReleaseRequestCtx(rctx)

		rctx.ctx, rctx.cancelFunc = context.WithCancel(ctx)
		defer rctx.cancelFunc()

		if request.TLS != nil {
			rctx.Request.flags.Add(_ReqFlagIsTLS)
		}
		if rctx.w == nil {
			rctx.w = bufio.NewWriterSize(writer, defaultRCtxPool.opt.SendBufferSize)
		} else {
			rctx.w.Reset(writer)
		}

		rctx.Request.fl1 = utils.B(request.Method)
		rctx.Request.fl2 = utils.B(request.RequestURI)
		rctx.Request.fl3 = utils.B(request.Proto)
		rctx.Request.Header().LoadMap(utils.MultiValueMap(request.Header))
		if request.Body != nil {
			rctx.Request.body = bodyBufPool.Get().(*bytes.Buffer)
			_, _ = io.Copy(rctx.Request.body, request.Body)
			_ = request.Body.Close()
		}

		defer func() {
			v := recover()
			if v != nil {
				writer.WriteHeader(StatusInternalServerError)
				log.Printf("sha: uncatched error, `%v`", v)
				return
			}

			res := &rctx.Response
			if res.statusCode != 0 {
				writer.WriteHeader(res.statusCode)
			} else {
				writer.WriteHeader(StatusOK)
			}
			res.Header().EachItem(func(item *utils.KvItem) bool {
				writer.Header().Add(utils.S(item.Key), utils.S(item.Val))
				return true
			})
			if res.cw != nil {
				_ = res.cw.Flush()
			}
			if res.body != nil {
				_, _ = io.Copy(writer, res.body)
			}
		}()
		handler.Handle(rctx)
	})
}
