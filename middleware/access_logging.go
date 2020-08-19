package middleware

import (
	"fmt"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/router"
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
	nfmt   *utils.NamedFmt
	logger *log.Logger
	opt    *AccessLoggingOption

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
	Enabled      bool
	TimeFmt      string
	DurationUnit time.Duration
}

// named formatter kwargs:
//
// 		Begin: begin time of the current request;
//
// 		End: end time of the current request;
//
// 		Cost: time spent processing current request
//
// 		ReqMethod: method of the current request
//
// 		ReqPath: path of the current request
//
//		ReqHeaders: all headers of the current request
//
//		ReqQuery: query of the current request
//
//		ReqForm: form body of the current request
//
//		ReqRemote: remote ip of the current request
//
//		ReqHeader@***: some header of the current request
//
//		ResStatusCode: status code of the current response
//
//		ResStatusText: status text of the current response
//
//		ResHeaders: all headers of the current response
//
//		ResBody: body of the current response
//
//		ResHeader@***: some header of the current response
//
//		ErrStack: error stack if an internal server error occurred
//
//		UserId: user id of the current request
func NewAccessLogger(fstr string, logger *log.Logger, opt *AccessLoggingOption) *_AccessLogger {
	rv := &_AccessLogger{
		nfmt:   utils.NewNamedFmt(fstr),
		logger: logger,
		opt:    opt,
	}

	for _, name := range rv.nfmt.Names {
		switch name {
		case "Begin":
			rv._beginT = true
		case "End":
			rv._endT = true
		case "Cost":
			rv._costT = true
		case "ReqMethod":
			rv._ReqMethod = true
		case "ReqPath":
			rv._ReqPath = true
		case "ReqHeaders":
			rv._ReqHeaders = true
		case "ReqQuery":
			rv._ReqQuery = true
		case "ReqForm":
			rv._ReqForm = true
		case "ReqRemote":
			rv._ReqRemote = true
		case "ResStatusCode":
			rv._ResStatusCode = true
		case "ResStatusText":
			rv._ResStatusText = true
		case "ResHeaders":
			rv._ResHeaders = true
		case "ResBody":
			rv._ResBody = true
		case "ErrStack":
			rv._ErrStack = true
		case "UserId":
			rv._UserId = true
		default:
			if strings.HasPrefix(name, "ReqHeader@") {
				rv._ReqHeader = append(rv._ReqHeader, name[10:])
			} else if strings.HasPrefix(name, "ResHeader@") {
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
		opt.TimeFmt = "2006-01-02 15:04:05.999999999"
	}
	return rv
}

type _Header interface {
	Peek(k string) []byte
	VisitAll(f func(key, value []byte))
}

func (al *_AccessLogger) peekAllHeader(header _Header) string {
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

func (al *_AccessLogger) peekSomeHeader(key string, hkeys []string, header _Header, m utils.M) {
	for _, hk := range hkeys {
		m[key+hk] = gotils.B2S(header.Peek(hk))
	}
}

func (al *_AccessLogger) peekRequest(m utils.M, ctx *fasthttp.RequestCtx) {
	if al._ReqMethod {
		m["ReqMethod"] = gotils.B2S(ctx.Method())
	}

	if al._ReqPath {
		m["ReqPath"] = gotils.B2S(ctx.Request.URI().Path())
	}

	if al._ReqRemote {
		m["ReqRemote"] = ctx.RemoteIP().String()
	}

	if al._ReqQuery {
		m["ReqQuery"] = ctx.QueryArgs().String()
	}

	if al._ReqForm {
		if ctx.PostArgs().Len() > 1 {
			m["ReqForm"] = ctx.PostArgs().String()
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
				m["ReqForm"] = buf.String()
			} else {
				m["ReqForm"] = ""
			}
		}
	}

	if al._ReqHeaders {
		m["ReqHeaders"] = al.peekAllHeader(&ctx.Request.Header)
	}

	if len(al._ReqHeader) > 0 {
		al.peekSomeHeader("ReqHeader@", al._ReqHeader, &ctx.Request.Header, m)
	}
}

func (al *_AccessLogger) peekResBody(ctx *fasthttp.RequestCtx) string {
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

func (al *_AccessLogger) peekResponse(m utils.M, ctx *fasthttp.RequestCtx) {
	if al._ResStatusCode {
		m["ResStatusCode"] = ctx.Response.StatusCode()
	}
	if al._ResStatusText {
		m["ResStatusText"] = http.StatusText(ctx.Response.StatusCode())
	}
	if al._ResBody {
		m["ResBody"] = al.peekResBody(ctx)
	}
	if al._ResHeaders {
		m["ResHeaders"] = al.peekAllHeader(&ctx.Response.Header)
	}

	if len(al._ResHeader) > 0 {
		al.peekSomeHeader("ResHeader@", al._ResHeader, &ctx.Response.Header, m)
	}
}

func (al *_AccessLogger) AsHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	if !al.opt.Enabled {
		return next
	}

	return func(ctx *fasthttp.RequestCtx) {
		if !al.opt.Enabled {
			next(ctx)
			return
		}

		var m = utils.M{}
		var begin time.Time
		if al._beginT {
			begin = time.Now()
		}

		al.peekRequest(m, ctx)

		defer func() {
			v := recover()
			var es string
			var end time.Time
			if al._endT {
				end = time.Now()
			}
			if al._costT {
				m["Cost"] = int64(end.Sub(begin) / al.opt.DurationUnit)
			}

			if v != nil {
				if al._ErrStack {
					switch val := v.(type) {
					case error:
						es = output.ErrorAndErrorStack(ctx, val)
					default:
						es = output.ErrorAndErrorStack(ctx, output.HttpErrors[fasthttp.StatusInternalServerError])
					}
					m["ErrStack"] = es
				} else {
					output.Recover(ctx, v)
				}
			} else if al._ErrStack {
				m["ErrStack"] = ""
			}

			al.peekResponse(m, ctx)

			if al._beginT {
				m["Begin"] = begin.Format(al.opt.TimeFmt)
			}

			if al._endT {
				m["End"] = end.Format(al.opt.TimeFmt)
			}

			if al._UserId {
				u, ok := auth.GetUser(ctx)
				if ok {
					m["UserId"] = u.GetId()
				} else {
					m["UserId"] = 0
				}
			}

			l := al.nfmt.Render(m)
			if al.logger == nil {
				log.Println(l)
			} else {
				al.logger.Println(l)
			}
		}()

		next(ctx)
	}
}

func (al *_AccessLogger) AsMiddleware() fasthttp.RequestHandler {
	return al.AsHandler(router.Next)
}
