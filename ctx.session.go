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
	byCookie bool
}

func (sra _SessionReqAdaptor) GetSessionID() *[]byte { return &(sra.ctx.session) }

func (sra _SessionReqAdaptor) SetSessionID() {
	if sra.byCookie {
		sra.ctx.Response.SetCookie(sessionOpts.CookieName, utils.S(sra.ctx.session), &sessionCookieOpts)
	} else {
		sra.ctx.Response.Header().Append(sessionOpts.HeaderName, sra.ctx.session)
		sra.ctx.Response.Header().AppendString(
			sessionOpts.HeaderMaxAgeName,
			strconv.FormatInt(int64(sessionOpts.MaxAge.Duration/time.Second), 10),
		)
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
var defaultSessionOpts = SessionOptions{
	Options:          session.DefaultOpts,
	CookieName:       "session",
	HeaderName:       "X-Session-ID",
	HeaderMaxAgeName: "X-Session-MaxAge",
}
var sessionCookieOpts CookieOptions
var sessionInitOnce sync.Once

func InitSession(backend session.Backend, opt *SessionOptions, cookieOpts *CookieOptions) {
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
			session.Init(&sessionOpts.Options, backend)
		},
	)
}

func (ctx *RequestCtx) Session() (session.Session, error) {
	if !ctx.sessionOK {
		byCookie := false
		sid, _ := ctx.Request.CookieValue(sessionOpts.CookieName)
		if len(sid) > 0 {
			byCookie = true
			ctx.session = append(ctx.session, sid...)
		} else {
			sid, _ = ctx.Request.Header().Get(sessionOpts.HeaderName)
			ctx.session = append(ctx.session, sid...)
		}
		return session.New(ctx, _SessionReqAdaptor{ctx, byCookie})
	}
	return ctx.session, nil
}
