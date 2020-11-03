package middleware

import (
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/router"
	"github.com/zzztttkkk/suna/utils"
	"log"
	"time"
)

func NewTimeoutAndAccessLoggingMiddleware(
	fstring string,
	timeoutDu time.Duration,
	timeoutCode int,
	timeoutMessage string,
	panicHandler func(ctx *fasthttp.RequestCtx, rv interface{}),
) router.Middleware {
	if len(fstring) < 0 {
		panic("suna.tal: empty format string")
	}

	logger := _NewAccessLogger(fstring)

	if timeoutDu > 0 {
		if timeoutCode < 1 {
			timeoutCode = fasthttp.StatusGatewayTimeout
		}

		if len(timeoutMessage) < 1 {
			timeoutMessage = fasthttp.StatusMessage(fasthttp.StatusGatewayTimeout)
		}

		statusText := fasthttp.StatusMessage(fasthttp.StatusGatewayTimeout)

		return router.MiddlewareFunc(
			func(ctx *fasthttp.RequestCtx, next func()) {
				handler := fasthttp.TimeoutWithCodeHandler(
					func(ctx *fasthttp.RequestCtx) {
						m := utils.M{}
						logger.peekRequest(m, ctx)

						defer func() { log.Print(logger.namedFmt.Render(m)) }()

						defer func() {
							isTimeout := ctx.LastTimeoutErrorResponse() != nil
							m["IsTimeout"] = isTimeout
							v := recover()
							if v != nil {
								logger.peekError(ctx, m, v)
								if !isTimeout {
									panicHandler(ctx, v)
								}
							}
							if isTimeout {
								m["TimeSpent"] = time.Now().Sub(ctx.Time()).Milliseconds()
								m["StatusCode"] = timeoutCode
								m["StatusText"] = statusText
							} else {
								logger.peekResponse(m, ctx)
							}
						}()

						next()
					},
					timeoutDu,
					timeoutMessage,
					timeoutCode,
				)

				handler(ctx)
			},
		)
	}

	return router.MiddlewareFunc(
		func(ctx *fasthttp.RequestCtx, next func()) {
			m := utils.M{}
			if logger._IsTimeout {
				m["IsTimeout"] = false
			}

			defer func() { log.Print(logger.namedFmt.Render(m)) }()

			defer func() {
				v := recover()
				if v != nil {
					logger.peekError(ctx, m, v)
					panicHandler(ctx, v)
				}
				logger.peekResponse(m, ctx)
			}()

			logger.peekRequest(m, ctx)
			next()
		},
	)
}
