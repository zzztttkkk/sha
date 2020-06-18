package middleware

import (
	"github.com/dchest/captcha"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/snow/ini"
	"github.com/zzztttkkk/snow/output"
	"github.com/zzztttkkk/snow/secret"
	"github.com/zzztttkkk/snow/utils"
	"time"
)

var bytesPool = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
var _t = []byte("0123456789")

func toString(digits []byte) string {
	s := make([]byte, len(digits), len(digits))
	for i, b := range digits {
		s[i] = _t[b]
	}
	return utils.B2s(s)
}

const (
	captchaIdKey   = "captcha.id"
	captchaUnixKey = "captcha.unix"
)

func (s *_SessionT) CaptchaGenerate(ctx *fasthttp.RequestCtx) {
	digits := secret.RandBytes(6, bytesPool)
	s.Set(captchaIdKey, toString(digits))
	s.Set(captchaIdKey, time.Now().Unix())

	ctx.Response.Header.Set("Cache-control", "no-store")
	ctx.Response.Header.Set("Content-type", "image/png")

	image := captcha.NewImage(s.key, digits, 480, 160)
	_, err := image.WriteTo(ctx)
	if err != nil {
		output.Error(ctx, err)
		return
	}
}

func (s *_SessionT) CaptchaVerify(ctx *fasthttp.RequestCtx) (ok bool) {
	if ini.IsDebug() {
		return true
	}

	defer func() {
		s.Del(captchaIdKey, captchaUnixKey)
		if !ok {
			output.StdError(ctx, fasthttp.StatusBadRequest)
		}
	}()

	ok = false
	v := ctx.FormValue("captcha")
	if v == nil {
		return
	}

	var code string
	if !s.Get(captchaIdKey, &code) {
		return
	}

	var unix int64
	if !s.Get(captchaUnixKey, &unix) || time.Now().Unix()-unix > 300 {
		return
	}

	if code != utils.B2s(v) {
		return
	}
	ok = true
	return
}
