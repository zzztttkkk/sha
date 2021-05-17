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

var CRSFTokenGenerator = func(v []byte) {
	const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	for i := 0; i < len(v); i++ {
		v[i] = encodeStd[rand.Int()%64]
	}
}

func init() {
	utils.MathRandSeed()
}

type SessionBackend interface {
	NewSession(ctx context.Context, session string, duration time.Duration) error
	ExistsSession(ctx context.Context, session string) bool
	ExpireSession(ctx context.Context, session string, duration time.Duration) error
	TTLSession(ctx context.Context, session string) time.Duration
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
	CookieName  string             `json:"cookie_name" toml:"cookie-name"`
	HeaderName  string             `json:"header_name" toml:"header-name"`
	KeyPrefix   string             `json:"key_prefix" toml:"key-prefix"`
	MaxAge      utils.TomlDuration `json:"max_age" toml:"max-age"`
	ServerSeqID uint64             `json:"server_id" toml:"server-id"`

	Captcha struct {
		ImageCaptcha captcha.Generator  `json:"-" toml:"-"`
		AudioCaptcha captcha.Generator  `json:"-" toml:"-"`
		MaxAge       utils.TomlDuration `json:"max_age" toml:"max-age"`
		Skip         bool               `json:"skip" toml:"skip"`
		seconds      int64
	} `json:"captcha" toml:"captcha"`

	CSRF struct {
		CookieName string             `json:"cookie_name" toml:"cookie-name"`
		HeaderName string             `json:"header_name" toml:"header-name"`
		MaxAge     utils.TomlDuration `json:"max_age" toml:"max-age"`
		seconds    int64
		Skip       bool `json:"skip" toml:"skip"`
	} `json:"csrf" toml:"csrf"`
}

var sessionBackend SessionBackend
var sessionOptions SessionOptions
var sessionCookieOptions CookieOptions
var once sync.Once

func UseSession(v SessionBackend, opt *SessionOptions) {
	once.Do(func() {
		sessionBackend = v
		sessionOptions = *opt
		sessionCookieOptions.MaxAge = int64(sessionOptions.MaxAge.Duration / time.Second)
		sessionOptions.Captcha.seconds = int64(sessionOptions.Captcha.MaxAge.Duration / time.Second)
		sessionOptions.CSRF.seconds = int64(sessionOptions.CSRF.MaxAge.Duration / time.Second)
	})
}

type Session []byte

func (s *Session) String() string { return utils.S(*s) }

func (s *Session) Key() string { return utils.S((*s)[len(sessionOptions.KeyPrefix)+1:]) }

func (s *Session) Set(ctx context.Context, key string, val interface{}) error {
	v, e := jsonx.Marshal(val)
	if e != nil {
		return e
	}
	return sessionBackend.SessionSet(ctx, s.Key(), key, v)
}

func (s *Session) Get(ctx context.Context, key string, dist interface{}) bool {
	v, e := sessionBackend.SessionGet(ctx, s.Key(), key)
	if e != nil {
		return false
	}
	return jsonx.Unmarshal(v, dist) == nil
}

func (s *Session) Del(ctx context.Context, keys ...string) error {
	return sessionBackend.SessionDel(ctx, s.Key(), keys...)
}

func (s *Session) GenerateImageCaptcha(ctx context.Context, w io.Writer) error {
	if sessionOptions.Captcha.ImageCaptcha == nil {
		return errors.New("sha: nil ImageCaptchaGenerator")
	}
	token, err := sessionOptions.Captcha.ImageCaptcha.GenerateTo(ctx, w)
	if err != nil {
		return err
	}
	_ = s.Set(ctx, ".captcha.token", token)
	_ = s.Set(ctx, ".captcha.created", time.Now().Unix())
	return nil
}

func (s *Session) GenerateAudioCaptcha(ctx context.Context, w io.Writer) error {
	if sessionOptions.Captcha.AudioCaptcha == nil {
		return errors.New("sha: nil ImageCaptchaGenerator")
	}
	token, err := sessionOptions.Captcha.AudioCaptcha.GenerateTo(ctx, w)
	if err != nil {
		return err
	}
	_ = s.Set(ctx, ".captcha.token", token)
	_ = s.Set(ctx, ".captcha.created", time.Now().Unix())
	return nil
}

func (s *Session) VerifyCaptcha(ctx context.Context, tokenInReq string) bool {
	if sessionOptions.Captcha.Skip {
		return true
	}
	if len(tokenInReq) < 1 {
		return false
	}
	var tokenInDB string
	var created int64
	s.Get(ctx, ".captcha.token", &tokenInDB)
	s.Get(ctx, ".captcha.created", &created)

	return tokenInDB == tokenInReq &&
		(sessionOptions.Captcha.seconds < 1 || time.Now().Unix()-created <= sessionOptions.Captcha.seconds)
}

func (s *Session) GenerateCSRFToken(ctx context.Context) string {
	var tmp = make([]byte, 16)
	CRSFTokenGenerator(tmp)
	_ = s.Set(ctx, ".crsf.token", tmp)
	_ = s.Set(ctx, ".crsf.created", time.Now().Unix())
	return utils.S(tmp)
}

func (s *Session) VerifyCRSFToken(ctx context.Context, token string) bool {
	if sessionOptions.CSRF.Skip {
		return true
	}
	var tokenInStorage string
	var created int64
	s.Get(ctx, ".crsf.token", &tokenInStorage)
	s.Get(ctx, ".crsf.created", &created)
	return tokenInStorage == token &&
		(sessionOptions.CSRF.seconds < 1 || time.Now().Unix()-created <= sessionOptions.CSRF.seconds)
}

func (ctx *RequestCtx) Session() (*Session, error) {
	if !ctx.sessionOK {
		ctx.session = append(ctx.session, sessionOptions.KeyPrefix...)
		ctx.session = append(ctx.session, ':')

		var sessionID []byte
		var byHeader bool
		var user auth.Subject
		user, _ = auth.Auth(ctx)
		if user != nil {
			var sid string
			if sessionBackend.Get(ctx, fmt.Sprintf("%s:auth:%d", sessionOptions.KeyPrefix, user.GetID()), &sid) {
				sessionID = append(sessionID, sid...)
			}
		}

		if len(sessionID) < 1 && len(sessionOptions.CookieName) > 0 {
			sessionID, _ = ctx.Request.CookieValue(sessionOptions.CookieName)
		}
		if len(sessionID) < 1 && len(sessionOptions.HeaderName) > 0 {
			sessionID, _ = ctx.Request.header.Get(sessionOptions.HeaderName)
			byHeader = true
		}

		if sessionBackend.ExistsSession(ctx, utils.S(sessionID)) {
			ctx.session = append(ctx.session, sessionID...)
		} else {
			// bad session id or session already expired
			SessionIDGenerator(&ctx.session)
			if e := sessionBackend.NewSession(ctx, ctx.session.Key(), sessionOptions.MaxAge.Duration); e != nil {
				return nil, e
			}
			if byHeader {
				ctx.Response.Header().SetString(sessionOptions.HeaderName, ctx.session.Key())
			} else {
				ctx.Response.SetCookie(sessionOptions.CookieName, ctx.session.Key(), &sessionCookieOptions)
			}
			if user != nil {
				_ = sessionBackend.Set(
					ctx,
					fmt.Sprintf("%sauth:%d", sessionOptions.KeyPrefix, user.GetID()),
					ctx.session.Key(), sessionOptions.MaxAge.Duration,
				)
			}
		}
		ctx.sessionOK = true
	}
	_ = sessionBackend.ExpireSession(ctx, ctx.session.Key(), sessionOptions.MaxAge.Duration)
	return &ctx.session, nil
}

type _RedisSessionBackend struct {
	cmd                redis.Cmdable
	hSetAndExpiresSha1 string
}

func (rsb *_RedisSessionBackend) TTLSession(ctx context.Context, session string) time.Duration {
	v, _ := rsb.cmd.TTL(ctx, session).Result()
	return v
}

var hSetAndExpiresScript = `redis.call('hset', KEYS[1], KEYS[2], ARGV[1]);redis.call('expire', KEYS[1], ARGV[2])`

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
	return rsb.cmd.EvalSha(
		ctx, rsb.hSetAndExpiresSha1,
		[]string{session, key},
		val, sessionOptions.MaxAge.Duration/time.Second,
	).Err()
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

func NewRedisSessionBackend(cmd redis.Cmdable) SessionBackend {
	hSetAndExpiresSha1, err := cmd.ScriptLoad(context.Background(), hSetAndExpiresScript).Result()
	if err != nil {
		panic(err)
	}
	return &_RedisSessionBackend{cmd: cmd, hSetAndExpiresSha1: hSetAndExpiresSha1}
}
