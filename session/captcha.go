package session

import (
	"github.com/dchest/captcha"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/secret"
	"time"
)

var bytesPool = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
var _t = []byte("0123456789")

func toString(digits []byte) string {
	s := make([]byte, len(digits), len(digits))
	for i, b := range digits {
		s[i] = _t[b]
	}
	return gotils.B2S(s)
}

const (
	captchaIdKey   = "~.suna.captcha.id"
	captchaUnixKey = "~.suna.captcha.unix"
)

var captchaWordSize int
var captchaWidth int
var captchaHeight int
var captchaForm string
var captchaMaxage int64

func _initCaptcha() {
	captchaWordSize = cfg.Session.Captcha.Words
	captchaHeight = cfg.Session.Captcha.Height
	captchaWidth = cfg.Session.Captcha.Width
	captchaForm = cfg.Session.Captcha.Form
	captchaMaxage = int64(cfg.Session.Captcha.Maxage)
}

func (ss Session) CaptchaGenerate(ctx *fasthttp.RequestCtx) {
	digits := secret.RandBytes(captchaWordSize, bytesPool)
	ss.Set(captchaIdKey, toString(digits))
	ss.Set(captchaIdKey, time.Now().Unix())

	ctx.Response.Header.Set("Cache-control", "no-store")
	ctx.Response.Header.Set("Content-type", "image/png")

	image := captcha.NewImage(string(ss), digits, captchaWidth, captchaHeight)
	_, err := image.WriteTo(output.NewCompressionWriter(ctx))
	if err != nil {
		output.Error(ctx, err)
		return
	}
}

func (ss Session) CaptchaVerify(ctx *fasthttp.RequestCtx) (ok bool) {
	if cfg.Session.Captcha.SkipInDebug && cfg.IsDebug() {
		return true
	}

	defer func() {
		ss.Del(captchaIdKey, captchaUnixKey) // del captcha anyway
		if !ok {
			output.Error(ctx, output.HttpErrors[fasthttp.StatusBadRequest])
		}
	}()

	ok = false
	v := ctx.FormValue(captchaForm)
	if v == nil {
		return
	}

	var code string
	if !ss.Get(captchaIdKey, &code) {
		return
	}

	var unix int64
	if !ss.Get(captchaUnixKey, &unix) || time.Now().Unix()-unix > captchaMaxage {
		return
	}

	if code != gotils.B2S(v) {
		return
	}
	ok = true
	return
}
