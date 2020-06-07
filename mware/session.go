package mware

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/rs/xid"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/router"
	"time"
)

var (
	SessionInCookie = "session"
	SessionInHeader = "session"
	SessionExpire   = time.Hour
)

type _SessionT struct {
	id   []byte
	User User
}

const sessionKey = "/s"

func GetSession(ctx *fasthttp.RequestCtx) *_SessionT {
	v := ctx.UserValue(sessionKey)
	if v == nil {
		return nil
	}
	s, ok := v.(*_SessionT)
	if !ok {
		return nil
	}
	return s
}

func setSession(ctx *fasthttp.RequestCtx, s *_SessionT) {
	ctx.SetUserValue(sessionKey, s)
}

var redisClient redis.Cmdable

func (s *_SessionT) cacheKey() string {
	return fmt.Sprintf("session:%s", s.id)
}

func (s *_SessionT) Set(key string, val interface{}) {
	bs, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}
	if err = redisClient.HSet(s.cacheKey(), key, bs).Err(); err != nil {
		panic(err)
	}
}

func (s *_SessionT) Get(key string, dst interface{}) bool {
	bs, err := redisClient.HGet(s.cacheKey(), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false
		}
		panic(err)
	}
	return json.Unmarshal(bs, dst) == nil
}

func (s *_SessionT) Del(keys ...string) {
	redisClient.HDel(s.cacheKey(), keys...)
}

func makeUSidKey(uid int64) string {
	return fmt.Sprintf("sonw:sessionid:%d", uid)
}

func SessionHandler(ctx *fasthttp.RequestCtx) {
	session := _SessionT{}
	var uid int64
	var sidKey string
	var user User

	user = GetUser(ctx)
	if user != nil {
		uid = user.Id()
		sidKey = makeUSidKey(uid)
		sid, _ := redisClient.Get(sidKey).Bytes()
		if len(sid) > 0 {
			session.id = sid
			redisClient.Set(sidKey, sid, SessionExpire)
			var sUid int64
			if !session.Get("uid", &sUid) || sUid != uid {
				// delete dirty data
				redisClient.Del(session.cacheKey(), sidKey)
				session.id = nil
			}
		}
	}

	if session.id == nil {
		session.id = ctx.Request.Header.Cookie(SessionInCookie)
		if session.id == nil {
			session.id = ctx.Request.Header.Peek(SessionInHeader)
			if session.id == nil {
				session.id = xid.New().Bytes()
				session.Set(".", 1)
			}
		}
	}

	if uid > 0 {
		session.Set("uid", uid)
	}

	redisClient.Expire(session.cacheKey(), SessionExpire)
	setSession(ctx, &session)

	ck := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(ck)
	ck.SetKey(SessionInCookie)
	ck.SetValueBytes(session.id)
	ck.SetPath("/")
	ck.SetMaxAge(int(SessionExpire / time.Second))
	ctx.Response.Header.SetCookie(ck)

	router.Next(ctx)
}
