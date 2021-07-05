package sha

import (
	"github.com/zzztttkkk/sha/session"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"sync"
	"time"
)

type _SessionReqAdaptor struct {
	ctx      *RequestCtx
	byHeader bool
}

func (sra _SessionReqAdaptor) UserAgent() string {
	ua, _ := sra.ctx.Request.Header().Get(HeaderUserAgent)
	return utils.S(ua)
}

func (sra _SessionReqAdaptor) GetSessionID() *[]byte { return &(sra.ctx.session) }

func (sra _SessionReqAdaptor) SetSessionID() {
	if sra.byHeader {
		sra.ctx.Response.Header().Append(sessionOpts.HeaderName, sra.ctx.session[session.PrefixLength:])
		sra.ctx.Response.Header().AppendString(
			sessionOpts.HeaderMaxAgeName,
			strconv.FormatInt(int64(sessionOpts.MaxAge.Duration/time.Second), 10),
		)
	} else {
		sra.ctx.Response.SetCookie(sessionOpts.CookieName, utils.S(sra.ctx.session[session.PrefixLength:]), &sessionCookieOpts)
	}
	sra.ctx.sessionOK = true
}

type SessionOptions struct {
	session.Options
	CookieName       string `toml:"cookie-name"`
	HeaderName       string `toml:"header-name"`
	HeaderMaxAgeName string `toml:"header-expires-name"`
}

var sessionOpts SessionOptions

var sessionCookieOpts CookieOptions
var sessionInitOnce sync.Once

func InitSession(opt *SessionOptions, cookieOpts *CookieOptions) {
	var defaultSessionOpts = SessionOptions{
		Options:          session.DefaultOpts,
		CookieName:       "session",
		HeaderName:       "X-Session-ID",
		HeaderMaxAgeName: "X-Session-MaxAge",
	}

	sessionInitOnce.Do(
		func() {
			if opt == nil {
				sessionOpts = defaultSessionOpts
			} else {
				sessionOpts = *opt
				utils.Merge(&sessionOpts, defaultSessionOpts)
			}

			if cookieOpts != nil {
				sessionCookieOpts = *cookieOpts
			}
			if sessionCookieOpts.MaxAge < 1 && sessionCookieOpts.Expires.IsZero() {
				sessionCookieOpts.MaxAge = int64(sessionOpts.MaxAge.Duration / time.Second)
			}
			session.Init(&sessionOpts.Options)
		},
	)
}

func (ctx *RequestCtx) Session() (session.Session, error) {
	if !ctx.sessionOK {
		if ctx.session == nil {
			ctx.session = make([]byte, 0, 25)
		}

		byHeader := false
		sid, _ := ctx.Request.CookieValue(sessionOpts.CookieName)
		if len(sid) > 0 {
			ctx.session = append(ctx.session, sid...)
		} else {
			sid, _ = ctx.Request.Header().Get(sessionOpts.HeaderName)
			if len(sid) > 0 {
				byHeader = true
				ctx.session = append(ctx.session, sid...)
			}
		}
		return session.New(ctx, _SessionReqAdaptor{ctx, byHeader})
	}
	return ctx.session, nil
}

func (ctx *RequestCtx) MustSession() session.Session {
	_, e := ctx.Session()
	if e != nil {
		panic(e)
	}
	return ctx.session
}
