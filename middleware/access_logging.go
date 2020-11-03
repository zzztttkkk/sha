package middleware

import (
	"fmt"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type _AccessLogger struct {
	namedFmt *utils.NamedFmt

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

	_IsTimeout bool

	_UserId bool
}

// revive:disable:cyclomatic
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
// 		reqHeaders: all headers of the current request
//
// 		query: query of the current request
//
// 		form: form body of the current request
//
// 		remote: remote ip of the current request
//
// 		reqHeader/***: some header of the current request
//
// 		statusCode: status code of the current response
//
// 		statusText: status text of the current response
//
// 		resHeaders: all headers of the current response
//
// 		resBody: body of the current response
//
// 		resHeader/***: some header of the current response
//
// 		errStack: error stack if an internal server error occurred
//
// 		userId: user id of the current request
func _NewAccessLogger(fstr string) *_AccessLogger {
	rv := &_AccessLogger{
		namedFmt: utils.NewNamedFmt(fstr),
	}

	for _, name := range rv.namedFmt.Names {
		switch name {
		case "Begin":
			rv._beginT = true
		case "End":
			rv._endT = true
		case "TimeSpent":
			rv._costT = true
		case "Method":
			rv._ReqMethod = true
		case "Path":
			rv._ReqPath = true
		case "ReqHeaders":
			rv._ReqHeaders = true
		case "Query":
			rv._ReqQuery = true
		case "Form":
			rv._ReqForm = true
		case "Remote":
			rv._ReqRemote = true
		case "StatusCode":
			rv._ResStatusCode = true
		case "StatusText":
			rv._ResStatusText = true
		case "ResHeaders":
			rv._ResHeaders = true
		case "ResBody":
			rv._ResBody = true
		case "ErrStack":
			rv._ErrStack = true
		case "UserId":
			rv._UserId = true
		case "IsTimeout":
			rv._IsTimeout = true
		default:
			if strings.HasPrefix(name, "ReqHeader/") {
				rv._ReqHeader = append(rv._ReqHeader, name[10:])
			} else if strings.HasPrefix(name, "ResHeader/") {
				rv._ResHeader = append(rv._ResHeader, name[10:])
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
	if logger._beginT {
		m["Begin"] = ctx.Time().Format(time.RFC1123Z)
	}

	if logger._ReqMethod {
		m["Method"] = gotils.B2S(ctx.Method())
	}

	if logger._ReqPath {
		m["Path"] = gotils.B2S(ctx.Request.URI().Path())
	}

	if logger._ReqRemote {
		m["Remote"] = ctx.RemoteIP().String()
	}

	if logger._ReqQuery {
		m["Query"] = ctx.QueryArgs().String()
	}

	if logger._UserId {
		u, ok := auth.GetUser(ctx)
		if ok {
			m["UserId"] = u.GetId()
		} else {
			m["UserId"] = 0
		}
	}

	if logger._ReqForm {
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

	if logger._ReqHeaders {
		m["ReqHeaders"] = logger.peekAllHeader(&ctx.Request.Header)
	}

	if len(logger._ReqHeader) > 0 {
		logger.peekSomeHeader("ReqHeader/", logger._ReqHeader, &ctx.Request.Header, m)
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
	if logger._ResStatusCode {
		m["StatusCode"] = ctx.Response.StatusCode()
	}
	if logger._ResStatusText {
		m["StatusText"] = http.StatusText(ctx.Response.StatusCode())
	}
	if logger._ResBody {
		m["ResBody"] = logger.peekResBody(ctx)
	}
	if logger._ResHeaders {
		m["ResHeaders"] = logger.peekAllHeader(&ctx.Response.Header)
	}

	if len(logger._ResHeader) > 0 {
		logger.peekSomeHeader("ResHeader/", logger._ResHeader, &ctx.Response.Header, m)
	}

	end := time.Now()

	if logger._endT {
		m["End"] = end.Format(time.RFC1123Z)
	}

	if logger._costT {
		m["TimeSpent"] = end.Sub(ctx.Time()).Milliseconds()
	}
}

func (logger *_AccessLogger) peekError(ctx *fasthttp.RequestCtx, m utils.M, v interface{}) {
	if logger._ErrStack {
		m["ErrStack"] = output.ErrorStack(v, 1)
	}
}
