package session

import (
	"context"
	"github.com/zzztttkkk/sha"
	"github.com/zzztttkkk/sha/utils"
	"io"
	"time"
)

type Request interface {
	GetSessionID() *[]byte
	SetSessionID()
}

type Session []byte

func (s *Session) String() string { return utils.S(*s) }

func (s *Session) Key() string { return utils.S((*s)[len(opts.Suffix)+1:]) }

func (s *Session) Set(ctx context.Context, key string, val interface{}) error {
	v, e := Marshal(val)
	if e != nil {
		return e
	}
	return backend.SessionSet(ctx, s.Key(), key, v)
}

func (s *Session) Get(ctx context.Context, key string, dist interface{}) bool {
	v, e := backend.SessionGet(ctx, s.Key(), key)
	if e != nil {
		return false
	}
	return Unmarshal(v, dist) == nil
}

func (s *Session) Del(ctx context.Context, keys ...string) error {
	return backend.SessionDel(ctx, s.Key(), keys...)
}

func (s *Session) Destroy(ctx context.Context) error { return backend.Del(ctx, utils.S(*s)) }

func (s *Session) Clear(ctx context.Context) error {
	return backend.ClearSession(ctx, utils.S(*s))
}

func (s *Session) GenerateImageCaptcha(ctx context.Context, w io.Writer) error {
	if ImageCaptchaGenerator == nil {
		return sha.StatusError(sha.StatusNotImplemented)
	}
	token, err := ImageCaptchaGenerator.GenerateTo(ctx, w)
	if err != nil {
		return err
	}
	_ = s.Set(ctx, ".captcha.token", token)
	_ = s.Set(ctx, ".captcha.created", time.Now().Unix())
	return nil
}

func (s *Session) GenerateAudioCaptcha(ctx context.Context, w io.Writer) error {
	if AudioCaptchaGenerator == nil {
		return sha.StatusError(sha.StatusNotImplemented)
	}
	token, err := AudioCaptchaGenerator.GenerateTo(ctx, w)
	if err != nil {
		return err
	}
	_ = s.Set(ctx, ".captcha.token", token)
	_ = s.Set(ctx, ".captcha.created", time.Now().Unix())
	return nil
}

func (s *Session) VerifyCaptcha(ctx context.Context, token string) bool {
	if opts.Captcha.Skip {
		return true
	}
	if len(token) < 1 {
		return false
	}
	var tokenInDB string
	var created int64
	s.Get(ctx, ".captcha.token", &tokenInDB)
	s.Get(ctx, ".captcha.created", &created)

	maxAge := int64(opts.Captcha.MaxAge.Duration / time.Second)
	return tokenInDB == token && (maxAge < 1 || time.Now().Unix()-created <= maxAge)
}

func (s *Session) GenerateCSRFToken(ctx context.Context) string {
	var tmp = make([]byte, 16)
	CRSFTokenGenerator(tmp)
	_ = s.Set(ctx, ".csrf.token", tmp)
	_ = s.Set(ctx, ".csrf.created", time.Now().Unix())
	return utils.S(tmp)
}

func (s *Session) VerifyCRSFToken(ctx context.Context, token string) bool {
	if opts.CSRF.Skip {
		return true
	}
	if len(token) != 16 {
		return false
	}

	var tokenInStorage string
	var created int64
	s.Get(ctx, ".csrf.token", &tokenInStorage)
	s.Get(ctx, ".csrf.created", &created)

	maxAge := int64(opts.CSRF.MaxAge.Duration / time.Second)
	return tokenInStorage == token && (maxAge < 1 || time.Now().Unix()-created <= maxAge)
}

func (s *Session) GetAll(ctx context.Context) map[string]string {
	return backend.SessionGetAll(ctx, utils.S(*s))
}

func (s *Session) Incr(ctx context.Context, key string, increment int64) (int64, error) {
	return backend.SessionIncr(ctx, utils.S(*s), key, increment)
}
