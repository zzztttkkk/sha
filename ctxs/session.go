package ctxs

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/rs/xid"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/utils"
	"strings"
	"time"
)

type SessionStorage string

var redisc redis.Cmdable
var sessionInCookie string
var sessionInHeader string
var sessionKeyPrefix string
var sessionExpire time.Duration

func _initSession() {
	redisc = cfg.RedisClient()

	sessionInCookie = cfg.Session.Cookie
	sessionInHeader = cfg.Session.Header
	sessionKeyPrefix = cfg.Session.Prefix
	if !strings.HasSuffix(sessionKeyPrefix, ":") {
		sessionKeyPrefix += ":"
	}
	sessionExpire = cfg.Session.MaxAge
	_initCaptcha()
}

func newSession(ctx *fasthttp.RequestCtx) SessionStorage {
	var sessionId string

	subject := User(ctx)
	if subject != nil {
		sidKey := fmt.Sprintf("%su:%d", sessionKeyPrefix, subject.GetId())
		sessionId = redisc.Get(sidKey).String()
		if len(sessionId) > 0 {
			sessionId = fmt.Sprintf("%s:%s", sessionKeyPrefix, sessionId)
			if redisc.Exists(sessionId).Val() == 1 {
				redisc.Expire(sidKey, sessionExpire)
				redisc.Expire(sessionId, sessionExpire)
				return SessionStorage(sessionId)
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
			return SessionStorage(sessionId)
		}
		sessionId = ""
	}

	now := time.Now()
	sessionId = fmt.Sprintf("%s:%s", sessionKeyPrefix, xid.New().String())
	redisc.HSet(sessionId, "cunix", now.Unix())
	if subject != nil {
		redisc.Set(fmt.Sprintf("%su:%d", sessionKeyPrefix, subject.GetId()), now.Unix(), sessionExpire)
	}

	return SessionStorage(sessionId)
}

func Session(ctx *fasthttp.RequestCtx) SessionStorage {
	ss, ok := ctx.UserValue(internal.RCtxKeySession).(SessionStorage)
	if ok {
		return ss
	}
	ss = newSession(ctx)
	ctx.SetUserValue(internal.RCtxKeySession, ss)
	return ss
}

func (ss SessionStorage) Get(key string, dist interface{}) bool {
	bs, err := redisc.HGet(string(ss), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false
		}
		panic(err)
	}
	return json.Unmarshal(bs, dist) == nil
}

func (ss SessionStorage) Set(key string, val interface{}) {
	bs, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}
	if err = redisc.HSet(string(ss), key, bs).Err(); err != nil {
		panic(err)
	}
}

func (ss SessionStorage) Del(keys ...string) {
	if err := redisc.HDel(string(ss), keys...).Err(); err != nil {
		if err == redis.Nil {
			return
		}
		panic(err)
	}
}
