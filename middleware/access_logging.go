package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/router"
	"github.com/zzztttkkk/suna/utils"
)

type AccessLogger struct {
	namedFmt *utils.NamedFmt
	logger   *log.Logger
	opt      *AccessLoggingOption

	_beginT bool
	_endT   bool
	_costT  bool

	_ReqMethod  bool
	_ReqPath    bool
	_ReqHeaders bool
	_ReqQuery   bool
	_ReqForm    bool
	_ReqRemote  bool

	_ResStatusCode bool
	_ResStatusText bool
	_ResHeaders    bool
	_ResBody       bool

	_ErrStack bool

	_ReqHeader []string
	_ResHeader []string

	_UserId bool
}

type AccessLoggingOption struct {
	TimeFmt         string
	DurationUnit    time.Duration
	AsGlobalRecover bool
}

//revive:disable:cyclomatic
// named formatter kwargs:
//
// 		begin: begin time of the current request;
//
// 		end: end time of the current request;
//
// 		cost: time spent processing current request
//
// 		method: method of the current request
//
// 		path: path of the current request
//
//		reqHeaders: all headers of the current request
//
//		query: query of the current request
//
//		form: form body of the current request
//
//		remote: remote ip of the current request
//
//		reqHeader/***: some header of the current request
//
//		statusCode: status code of the current response
//
//		statusText: status text of the current response
//
//		resHeaders: all headers of the current response
//
//		resBody: body of the current response
//
//		resHeader/***: some header of the current response
//
//		errStack: error stack if an internal server error occurred
//
//		userId: user id of the current request
func NewAccessLogger(fstr string, logger *log.Logger, opt *AccessLoggingOption) *AccessLogger {
	if opt == nil {
		opt = &AccessLoggingOption{}
	}

	if opt.DurationUnit == 0 {
		opt.DurationUnit = time.Millisecond
	}

	rv := &AccessLogger{
		namedFmt: utils.NewNamedFmt(fstr),
		logger:   logger,
		opt:      opt,
	}

	for _, name := range rv.namedFmt.Names {
		switch name {
		case "begin":
			rv._beginT = true
		case "end":
			rv._endT = true
		case "cost":
			rv._costT = true
		case "method":
			rv._ReqMethod = true
		case "path":
			rv._ReqPath = true
		case "reqHeaders":
			rv._ReqHeaders = true
		case "query":
			rv._ReqQuery = true
		case "form":
			rv._ReqForm = true
		case "remote":
			rv._ReqRemote = true
		case "statusCode":
			rv._ResStatusCode = true
		case "statusText":
			rv._ResStatusText = true
		case "resHeaders":
			rv._ResHeaders = true
		case "resBody":
			rv._ResBody = true
		case "errStack":
			rv._ErrStack = true
		case "userId":
			rv._UserId = true
		default:
			if strings.HasPrefix(name, "reqHeader/") {
				rv._ReqHeader = append(rv._ReqHeader, name[10:])
			} else if strings.HasPrefix(name, "resHeader/") {
				rv._ResHeader = append(rv._ResHeader, name[10:])
			} else {
				panic(fmt.Errorf("suna.middleware.access_logging: unknown name `%s`", name))
			}
		}
	}

	if rv._costT {
		rv._endT = true
		rv._beginT = true
	}

	if len(opt.TimeFmt) < 1 {
		opt.TimeFmt = "2006-01-02 15:04:05"
	}
	return rv
}

type _Header interface {
	Peek(k string) []byte
	VisitAll(f func(key, value []byte))
}

func (al *AccessLogger) peekAllHeader(header _Header) string {
	buf := strings.Builder{}
	header.VisitAll(
		func(key, value []byte) {
			buf.Write(key)
			buf.WriteString("=")
			buf.Write(value)
			buf.WriteString("; ")
		},
	)
	return buf.String()
}

func (al *AccessLogger) peekSomeHeader(key string, hkeys []string, header _Header, m utils.M) {
	for _, hk := range hkeys {
		m[key+hk] = gotils.B2S(header.Peek(hk))
	}
}

func (al *AccessLogger) peekRequest(m utils.M, ctx *fasthttp.RequestCtx) {
	if al._ReqMethod {
		m["method"] = gotils.B2S(ctx.Method())
	}

	if al._ReqPath {
		m["path"] = gotils.B2S(ctx.Request.URI().Path())
	}

	if al._ReqRemote {
		m["remote"] = ctx.RemoteIP().String()
	}

	if al._ReqQuery {
		m["query"] = ctx.QueryArgs().String()
	}

	if al._ReqForm {
		if ctx.PostArgs().Len() > 1 {
			m["form"] = ctx.PostArgs().String()
		} else {
			mf, e := ctx.MultipartForm()
			if e == nil {
				buf := strings.Builder{}
				for k, vl := range mf.Value {
					buf.WriteString(k)
					buf.WriteString("=[")
					buf.WriteString(strings.Join(vl, ", "))
					buf.WriteString("]")
				}
				for k, vl := range mf.File {
					buf.WriteString(k)
					buf.WriteString("<file>=[")
					for _, fh := range vl {
						buf.WriteString(fh.Filename)
						buf.WriteString(strconv.FormatInt(fh.Size, 10))
						buf.WriteString(", ")
					}
					buf.WriteString("]")
				}
				m["form"] = buf.String()
			} else {
				m["form"] = ""
			}
		}
	}

	if al._ReqHeaders {
		m["reqHeaders"] = al.peekAllHeader(&ctx.Request.Header)
	}

	if len(al._ReqHeader) > 0 {
		al.peekSomeHeader("reqHeader/", al._ReqHeader, &ctx.Request.Header, m)
	}
}

func (al *AccessLogger) peekResBody(ctx *fasthttp.RequestCtx) string {
	switch gotils.B2S(ctx.Response.Header.Peek("Content-Encoding")) {
	case "deflate":
		d, _ := ctx.Response.BodyInflate()
		return gotils.B2S(d)
	case "gzip":
		d, _ := ctx.Response.BodyGunzip()
		return gotils.B2S(d)
	default:
		return gotils.B2S(ctx.Response.Body())
	}
}

func (al *AccessLogger) peekResponse(m utils.M, ctx *fasthttp.RequestCtx) {
	if al._ResStatusCode {
		m["statusCode"] = ctx.Response.StatusCode()
	}
	if al._ResStatusText {
		m["statusText"] = http.StatusText(ctx.Response.StatusCode())
	}
	if al._ResBody {
		m["resBody"] = al.peekResBody(ctx)
	}
	if al._ResHeaders {
		m["resHeaders"] = al.peekAllHeader(&ctx.Response.Header)
	}

	if len(al._ResHeader) > 0 {
		al.peekSomeHeader("resHeader/", al._ResHeader, &ctx.Response.Header, m)
	}
}

func (al *AccessLogger) AsHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var m = utils.M{}
		var begin time.Time
		if al._beginT {
			begin = time.Now()
		}

		al.peekRequest(m, ctx)

		defer func() {
			var v interface{}
			if al._ErrStack {
				v = recover()
				if v != nil {
					m["errStack"] = output.ErrorStack(v, 1)
					if al.opt.AsGlobalRecover {
						output.Error(ctx, v)
					}
				}
			}

			var end time.Time
			if al._endT {
				end = time.Now()
			}
			if al._costT {
				m["cost"] = int64(end.Sub(begin) / al.opt.DurationUnit)
			}

			if v == nil || al.opt.AsGlobalRecover {
				al.peekResponse(m, ctx)
			}

			if al._beginT {
				m["begin"] = begin.Format(al.opt.TimeFmt)
			}

			if al._endT {
				m["end"] = end.Format(al.opt.TimeFmt)
			}

			if al._UserId {
				u, ok := auth.GetUser(ctx)
				if ok {
					m["userId"] = u.GetId()
				} else {
					m["userId"] = 0
				}
			}

			l := al.namedFmt.Render(m)
			if al.logger == nil {
				log.Print(l)
			} else {
				al.logger.Print(l)
			}

			if v != nil && !al.opt.AsGlobalRecover {
				panic(v)
			}
		}()

		next(ctx)
	}
}

func (al *AccessLogger) AsMiddleware() fasthttp.RequestHandler {
	return al.AsHandler(router.Next)
}
