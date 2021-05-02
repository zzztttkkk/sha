package sha

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/zzztttkkk/sha/auth"
	"github.com/zzztttkkk/sha/captcha"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"io"
	"math/bits"
	"math/rand"
	"sync"
	"time"
)

var SessionIDGenerator = func(v *Session) {
	const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	writeUint := func(q uint64) {
		var r uint64
		for {
			q, r = bits.Div64(0, q, 64)
			*v = append(*v, encodeStd[r])
			if q == 0 {
				break
			}
		}
	}
	writeUint(sessionOptions.ServerSeqID)
	*v = append(*v, '.')
	writeUint(uint64(time.Now().UnixNano()))
	*v = append(*v, '.')
	for i := 0; i < 6; i++ {
		*v = append(*v, encodeStd[rand.Int()%64])
	}
}

func init() {
	utils.MathRandSeed()
}

type SessionBackend interface {
	NewSession(ctx context.Context, session string, duration time.Duration) error
	ExistsSession(ctx context.Context, session string) bool
	ExpireSession(ctx context.Context, session string, duration time.Duration) error
	RemoveSession(ctx context.Context, session string) error

	SessionSet(ctx context.Context, session string, key string, val []byte) error
	SessionGet(ctx context.Context, session string, key string) ([]byte, error)
	SessionDel(ctx context.Context, session string, keys ...string) error
	SessionGetAll(ctx context.Context, session string) map[string]string

	Set(ctx context.Context, key string, val interface{}, duration time.Duration) error
	Get(ctx context.Context, key string, dist interface{}) bool
	Del(ctx context.Context, keys ...string) error
}

type SessionOptions struct {
	CookieName   string             `json:"cookie_name" toml:"cookie-name"`
	HeaderName   string             `json:"header_name" toml:"header-name"`
	KeyPrefix    string             `json:"key_prefix" toml:"key-prefix"`
	MaxAge       utils.TomlDuration `json:"max_age" toml:"max-age"`
	Auth         bool               `json:"auth" toml:"auth"`
	ServerSeqID  uint64             `json:"server_seq_id" toml:"server-seq-id"`
	ImageCaptcha captcha.Generator  `json:"-" toml:"-"`
	AudioCaptcha captcha.Generator  `json:"-" toml:"-"`
}

var defaultSessionOptions = SessionOptions{}
var sessionBackend SessionBackend
var sessionOptions SessionOptions
var sessionCookieOptions *CookieOptions
var sessionImageCaptcha captcha.Generator
var sessionAudioCaptcha captcha.Generator
var once sync.Once

func UseSession(v SessionBackend, opt *SessionOptions) {
	once.Do(func() {
		sessionBackend = v
		if opt == nil {
			opt = &defaultSessionOptions
		}
		sessionOptions = *opt
		sessionCookieOptions = &CookieOptions{MaxAge: int64(sessionOptions.MaxAge.Duration / time.Second)}
		sessionAudioCaptcha = opt.AudioCaptcha
		sessionImageCaptcha = opt.ImageCaptcha
	})
}

type Session []byte

func (s *Session) String() string { return utils.S((*s)[len(sessionOptions.KeyPrefix):]) }

func (s *Session) key() string { return utils.S(*s) }

func (s *Session) Set(ctx context.Context, key string, val interface{}) error {
	v, e := jsonx.Marshal(val)
	if e != nil {
		return e
	}
	return sessionBackend.SessionSet(ctx, s.key(), key, v)
}

func (s *Session) Get(ctx context.Context, key string, dist interface{}) bool {
	v, e := sessionBackend.SessionGet(ctx, s.key(), key)
	if e != nil {
		return false
	}
	return jsonx.Unmarshal(v, dist) == nil
}

func (s *Session) Del(ctx context.Context, keys ...string) error {
	return sessionBackend.SessionDel(ctx, s.key(), keys...)
}

func (s *Session) ImageCaptcha(ctx context.Context, w io.Writer) error {
	if sessionImageCaptcha == nil {
		return errors.New("sha: nil ImageCaptchaGenerator")
	}
	token, err := sessionImageCaptcha.GenerateTo(ctx, w)
	if err != nil {
		return err
	}
	_ = s.Set(ctx, "captcha.token", token)
	_ = s.Set(ctx, "captcha.created", time.Now().Unix())
	return nil
}

func (s *Session) AudioCaptcha(ctx context.Context, w io.Writer) error {
	if sessionAudioCaptcha == nil {
		return errors.New("sha: nil ImageCaptchaGenerator")
	}
	token, err := sessionAudioCaptcha.GenerateTo(ctx, w)
	if err != nil {
		return err
	}
	_ = s.Set(ctx, "captcha.token", token)
	_ = s.Set(ctx, "captcha.created", time.Now().Unix())
	return nil
}

func (s *Session) CaptchaVerify(ctx context.Context, tokenInReq string) bool {
	if len(tokenInReq) < 1 {
		return false
	}
	var tokenInDB string
	var created int64
	s.Get(ctx, "captcha.token", &tokenInDB)
	s.Get(ctx, "captcha.created", &created)
	return time.Now().Unix()-created < 300 && tokenInDB == tokenInReq
}

func (ctx *RequestCtx) Session() (*Session, error) {
	if !ctx.sessionOK {
		ctx.session = append(ctx.session, sessionOptions.KeyPrefix...)

		var sessionID []byte
		var byHeader bool
		var user auth.Subject
		if sessionOptions.Auth {
			user, _ = auth.Auth(ctx)
			if user != nil {
				var sid string
				if sessionBackend.Get(ctx, fmt.Sprintf("%sauth:%d", sessionOptions.KeyPrefix, user.GetID()), &sid) {
					sessionID = append(sessionID, sid...)
				}
			}
		}

		if len(sessionOptions.CookieName) > 0 {
			sessionID, _ = ctx.Request.CookieValue(sessionOptions.CookieName)
		}
		if len(sessionID) < 0 && len(sessionOptions.HeaderName) > 0 {
			sessionID, _ = ctx.Request.header.Get(sessionOptions.HeaderName)
			byHeader = true
		}

		if sessionBackend.ExistsSession(ctx, utils.S(sessionID)) {
			ctx.session = append(ctx.session, sessionID...)
		} else {
			// bad session id or session already expired
			SessionIDGenerator(&ctx.session)
			if e := sessionBackend.NewSession(ctx, ctx.session.key(), sessionOptions.MaxAge.Duration); e != nil {
				return nil, e
			}
			if byHeader {
				ctx.Response.Header().SetString(sessionOptions.HeaderName, ctx.session.String())
			} else {
				ctx.Response.SetCookie(sessionOptions.CookieName, ctx.session.String(), sessionCookieOptions)
			}
			if user != nil {
				_ = sessionBackend.Set(
					ctx,
					fmt.Sprintf("%sauth:%d", sessionOptions.KeyPrefix, user.GetID()),
					ctx.session.String(), sessionOptions.MaxAge.Duration,
				)
			}
		}
		ctx.sessionOK = true
	}
	_ = sessionBackend.ExpireSession(ctx, ctx.session.key(), sessionOptions.MaxAge.Duration)
	return &ctx.session, nil
}

type _RedisSessionBackend struct {
	cmd redis.Cmdable
}

func (rsb *_RedisSessionBackend) Set(ctx context.Context, key string, val interface{}, duration time.Duration) error {
	v, e := jsonx.Marshal(val)
	if e != nil {
		return e
	}
	return rsb.cmd.Set(ctx, key, v, duration).Err()
}

func (rsb *_RedisSessionBackend) Get(ctx context.Context, key string, dist interface{}) bool {
	v, e := rsb.cmd.Get(ctx, key).Bytes()
	if e != nil {
		return false
	}
	return jsonx.Unmarshal(v, dist) == nil
}

func (rsb *_RedisSessionBackend) Del(ctx context.Context, keys ...string) error {
	return rsb.cmd.Del(ctx, keys...).Err()
}

func (rsb *_RedisSessionBackend) SessionGetAll(ctx context.Context, session string) map[string]string {
	v, _ := rsb.cmd.HGetAll(ctx, session).Result()
	return v
}

func (rsb *_RedisSessionBackend) ExistsSession(ctx context.Context, session string) bool {
	v, _ := rsb.cmd.Exists(ctx, session).Result()
	return v == 1
}

func (rsb *_RedisSessionBackend) NewSession(ctx context.Context, session string, duration time.Duration) error {
	if err := rsb.cmd.HSet(ctx, session, ".created", time.Now().Unix()).Err(); err != nil {
		return err
	}
	return rsb.cmd.Expire(ctx, session, duration).Err()
}

func (rsb *_RedisSessionBackend) ExpireSession(ctx context.Context, session string, duration time.Duration) error {
	return rsb.cmd.Expire(ctx, session, duration).Err()
}

func (rsb *_RedisSessionBackend) SessionSet(ctx context.Context, session string, key string, val []byte) error {
	return rsb.cmd.HSet(ctx, session, key, val).Err()
}

func (rsb *_RedisSessionBackend) SessionGet(ctx context.Context, session string, key string) ([]byte, error) {
	return rsb.cmd.HGet(ctx, session, key).Bytes()
}

func (rsb *_RedisSessionBackend) SessionDel(ctx context.Context, session string, keys ...string) error {
	return rsb.cmd.HDel(ctx, session, keys...).Err()
}

func (rsb *_RedisSessionBackend) RemoveSession(ctx context.Context, session string) error {
	return rsb.cmd.Del(ctx, session).Err()
}

var _ SessionBackend = (*_RedisSessionBackend)(nil)

func NewRedisSessionBackend(cmd redis.Cmdable) SessionBackend { return &_RedisSessionBackend{cmd: cmd} }
