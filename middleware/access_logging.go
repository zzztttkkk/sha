package middleware

import (
	"fmt"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type _AccessLogger struct {
	namedFmt *utils.NamedFmt

	beginTime bool
	endTime   bool
	spentTime bool

	reqMethod  bool
	reqPath    bool
	reqHeaders bool
	reqQuery   bool
	reqForm    bool
	reqRemote  bool

	resStatusCode bool
	resStatusText bool
	resHeaders    bool
	resBody       bool

	errStack bool

	reqHeader []string
	resHeader []string

	isTimeout bool

	userId bool

	panicHandler func(ctx *fasthttp.RequestCtx, v interface{})
}

// named formatter kwargs:
// 		Begin: begin time of the current request;
// 		End: end time of the current request;
// 		TimeSpent: time spent processing current request
// 		Method: method of the current request
// 		Path: path of the current request
// 		ReqHeaders: all headers of the current request
// 		Query: query of the current request
// 		Form: form body of the current request
// 		Remote: remote ip of the current request
// 		ReqHeader/***: some header of the current request
// 		StatusCode: status code of the current response
// 		StatusText: status text of the current response
// 		ResHeaders: all headers of the current response
// 		ResBody: body of the current response
// 		ResHeader/***: some header of the current response
// 		ErrStack: error stack if an internal server error occurred
// 		UserId: user id of the current request
//		IsTimeout: the request processing timeout
func NewAccessLogger(fstr string, panicHandler func(ctx *fasthttp.RequestCtx, v interface{})) *_AccessLogger {
	rv := &_AccessLogger{
		namedFmt:     utils.NewNamedFmt(fstr),
		panicHandler: panicHandler,
	}

	for _, name := range rv.namedFmt.Names {
		switch name {
		case "Begin":
			rv.beginTime = true
		case "End":
			rv.endTime = true
		case "TimeSpent":
			rv.spentTime = true
		case "Method":
			rv.reqMethod = true
		case "Path":
			rv.reqPath = true
		case "ReqHeaders":
			rv.reqHeaders = true
		case "Query":
			rv.reqQuery = true
		case "Form":
			rv.reqForm = true
		case "Remote":
			rv.reqRemote = true
		case "StatusCode":
			rv.resStatusCode = true
		case "StatusText":
			rv.resStatusText = true
		case "ResHeaders":
			rv.resHeaders = true
		case "ResBody":
			rv.resBody = true
		case "ErrStack":
			rv.errStack = true
		case "UserId":
			rv.userId = true
		case "IsTimeout":
			rv.isTimeout = true
		default:
			if strings.HasPrefix(name, "ReqHeader/") {
				rv.reqHeader = append(rv.reqHeader, name[10:])
			} else if strings.HasPrefix(name, "ResHeader/") {
				rv.resHeader = append(rv.resHeader, name[10:])
			} else {
				panic(fmt.Errorf("suna.middleware.access_logging: unknown name `%s`", name))
			}
		}
	}
	return rv
}

type _Header interface {
	Peek(k string) []byte
	VisitAll(f func(key, value []byte))
}

func (logger *_AccessLogger) peekAllHeader(header _Header) string {
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

func (logger *_AccessLogger) peekSomeHeader(key string, hkeys []string, header _Header, m utils.M) {
	for _, hk := range hkeys {
		m[key+hk] = gotils.B2S(header.Peek(hk))
	}
}

func (logger *_AccessLogger) peekRequest(m utils.M, ctx *fasthttp.RequestCtx) {
	if logger.beginTime {
		m["Begin"] = ctx.Time().Format(time.RFC1123Z)
	}

	if logger.reqMethod {
		m["Method"] = gotils.B2S(ctx.Method())
	}

	if logger.reqPath {
		m["Path"] = gotils.B2S(ctx.Request.URI().Path())
	}

	if logger.reqRemote {
		m["Remote"] = ctx.RemoteIP().String()
	}

	if logger.reqQuery {
		m["Query"] = ctx.QueryArgs().String()
	}

	if logger.userId {
		u, ok := auth.GetUser(ctx)
		if ok {
			m["UserId"] = u.GetId()
		} else {
			m["UserId"] = 0
		}
	}

	if logger.reqForm {
		if ctx.PostArgs().Len() > 1 {
			m["Form"] = ctx.PostArgs().String()
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
					buf.WriteString("<file>[")
					for _, fh := range vl {
						buf.WriteString(fh.Filename)
						buf.WriteString(strconv.FormatInt(fh.Size, 10))
						buf.WriteString(", ")
					}
					buf.WriteString("]")
				}
				m["Form"] = buf.String()
			} else {
				m["Form"] = ""
			}
		}
	}

	if logger.reqHeaders {
		m["ReqHeaders"] = logger.peekAllHeader(&ctx.Request.Header)
	}

	if len(logger.reqHeader) > 0 {
		logger.peekSomeHeader("ReqHeader/", logger.reqHeader, &ctx.Request.Header, m)
	}
}

func (logger *_AccessLogger) peekResBody(ctx *fasthttp.RequestCtx) string {
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

func (logger *_AccessLogger) peekResponse(m utils.M, ctx *fasthttp.RequestCtx) {
	if logger.resStatusCode {
		m["StatusCode"] = ctx.Response.StatusCode()
	}
	if logger.resStatusText {
		m["StatusText"] = http.StatusText(ctx.Response.StatusCode())
	}
	if logger.resBody {
		m["ResBody"] = logger.peekResBody(ctx)
	}
	if logger.resHeaders {
		m["ResHeaders"] = logger.peekAllHeader(&ctx.Response.Header)
	}

	if len(logger.resHeader) > 0 {
		logger.peekSomeHeader("ResHeader/", logger.resHeader, &ctx.Response.Header, m)
	}

	end := time.Now()

	if logger.endTime {
		m["End"] = end.Format(time.RFC1123Z)
	}

	if logger.spentTime {
		m["TimeSpent"] = end.Sub(ctx.Time()).Milliseconds()
	}
}

func (logger *_AccessLogger) peekError(ctx *fasthttp.RequestCtx, m utils.M, v interface{}) {
	if logger.errStack {
		m["ErrStack"] = output.ErrorStack(v, 1)
	}
}

func (logger *_AccessLogger) Process(ctx *fasthttp.RequestCtx, next func()) {
	m := utils.M{}
	logger.peekRequest(m, ctx)

	defer func() { log.Print(logger.namedFmt.Render(m)) }()

	defer func() {
		if logger.errStack {
			if v := recover(); v != nil {
				logger.peekError(ctx, m, v)
				if logger.panicHandler != nil {
					logger.panicHandler(ctx, v)
				}
			}
		}
		logger.peekResponse(m, ctx)
	}()

	next()
}
