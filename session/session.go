package session

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/rs/xid"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/auth"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/utils"
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
	sessionExpire = cfg.Session.MaxAge
	_initCaptcha()
}

func newSession(ctx *fasthttp.RequestCtx) Session {
	var sessionId string

	subject := auth.GetUser(ctx)
	if subject != nil {
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
			sessionId = utils.B2s(sv)
		}
	}

	if len(sessionId) < 1 {
		sv := ctx.Request.Header.Peek(sessionInHeader)
		if sv != nil {
			sessionId = utils.B2s(sv)
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

func (ss Session) Get(key string, dist interface{}) bool {
	bs, err := redisc.HGet(string(ss), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false
		}
		panic(err)
	}
	return json.Unmarshal(bs, dist) == nil
}

func (ss Session) Set(key string, val interface{}) {
	bs, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}
	if err = redisc.HSet(string(ss), key, bs).Err(); err != nil {
		panic(err)
	}
}

func (ss Session) Del(keys ...string) {
	if err := redisc.HDel(string(ss), keys...).Err(); err != nil {
		if err == redis.Nil {
			return
		}
		panic(err)
	}
}

func (ss Session) Refresh() {
	redisc.Expire(string(ss), sessionExpire)
}

func Get(ctx *fasthttp.RequestCtx) Session {
	return newSession(ctx)
}

func GetMust(ctx *fasthttp.RequestCtx) (s Session) {
	if s = newSession(ctx); len(s) < 1 {
		panic(output.HttpErrors[fasthttp.StatusForbidden])
	}
	return
}
