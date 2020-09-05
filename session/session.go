package session

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/rs/xid"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/jsonx"
	"github.com/zzztttkkk/suna/output"
	"strings"
	"time"
)

type Session string

var sessionInCookie string
var sessionInHeader string
var sessionKeyPrefix string
var sessionExpire time.Duration

func _initSession() {
	sessionInCookie = cfg.Session.Cookie
	sessionInHeader = cfg.Session.Header
	sessionKeyPrefix = cfg.Session.Prefix
	if !strings.HasSuffix(sessionKeyPrefix, ":") {
		sessionKeyPrefix += ":"
	}
	sessionExpire = cfg.Session.Maxage.Duration
	_initCaptcha()
}

func newSession(ctx *fasthttp.RequestCtx) Session {
	var sessionId string

	subject, ok := auth.GetUser(ctx)
	if ok {
		sidKey := fmt.Sprintf("%su:%d", sessionKeyPrefix, subject.GetId())
		sessionId = redisc.Get(sidKey).String()
		if len(sessionId) > 0 {
			sessionId = fmt.Sprintf("%s:%s", sessionKeyPrefix, sessionId)
			if redisc.Exists(sessionId).Val() == 1 {
				redisc.Expire(sidKey, sessionExpire)
				redisc.Expire(sessionId, sessionExpire)
				return Session(sessionId)
			}
			redisc.Del(sidKey)
			sessionId = ""
		}
	}

	if len(sessionInCookie) > 0 {
		sv := ctx.Request.Header.Cookie(sessionInCookie)
		if sv != nil {
			sessionId = gotils.B2S(sv)
		}
	}

	if len(sessionId) < 1 {
		sv := ctx.Request.Header.Peek(sessionInHeader)
		if sv != nil {
			sessionId = gotils.B2S(sv)
		}
	}

	if len(sessionId) > 1 {
		sessionId = fmt.Sprintf("%s:%s", sessionKeyPrefix, sessionId)
		if redisc.Exists(sessionId).Val() == 1 {
			redisc.Expire(sessionId, sessionExpire)
			return Session(sessionId)
		}
		sessionId = ""
	}

	now := time.Now()
	sessionId = fmt.Sprintf("%s:%s", sessionKeyPrefix, xid.New().String())
	redisc.HSet(sessionId, internal.SessionExistsKey, 1)
	if subject != nil {
		redisc.Set(fmt.Sprintf("%su:%d", sessionKeyPrefix, subject.GetId()), now.Unix(), sessionExpire)
	}

	return Session(sessionId)
}

func (sion Session) Get(key string, dist interface{}) bool {
	bs, err := redisc.HGet(string(sion), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false
		}
		panic(err)
	}
	return jsonx.Unmarshal(bs, dist) == nil
}

func (sion Session) Set(key string, val interface{}) {
	bs, err := jsonx.Marshal(val)
	if err != nil {
		panic(err)
	}
	if err = redisc.HSet(string(sion), key, bs).Err(); err != nil {
		panic(err)
	}
}

func (sion Session) Del(keys ...string) {
	if err := redisc.HDel(string(sion), keys...).Err(); err != nil {
		if err == redis.Nil {
			return
		}
		panic(err)
	}
}

func (sion Session) Refresh() {
	redisc.Expire(string(sion), sessionExpire)
}

func New(ctx *fasthttp.RequestCtx) (s Session) {
	si := ctx.UserValue(internal.RCtxSessionKey)
	if si != nil {
		return si.(Session)
	}
	if s = newSession(ctx); len(s) < 1 {
		panic(output.HttpErrors[fasthttp.StatusForbidden])
	}
	ctx.SetUserValue(internal.RCtxSessionKey, s)
	return
}
