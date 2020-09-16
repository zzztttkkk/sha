package session

import (
	"time"

	"github.com/dchest/captcha"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/secret"
)

var bytesPool = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
var _t = []byte("0123456789")

func toString(digits []byte) string {
	s := make([]byte, len(digits))
	for i, b := range digits {
		s[i] = _t[b]
	}
	return gotils.B2S(s)
}

var captchaWordSize int
var captchaWidth int
var captchaHeight int
var captchaForm string
var captchaMaxage int64
var skipVerify bool
var captchaAudioLang string

func _initCaptcha() {
	captchaWordSize = cfg.Session.Captcha.TokenSize
	captchaHeight = cfg.Session.Captcha.ImageHeight
	captchaWidth = cfg.Session.Captcha.ImageWidth
	captchaForm = cfg.Session.Captcha.Form
	captchaMaxage = int64(cfg.Session.Captcha.Maxage)
	skipVerify = cfg.IsDebug() && cfg.Session.SkipVerifyInDebug
	captchaAudioLang = cfg.Session.Captcha.AudioLanguage
}

func (sion Session) CaptchaGenerateImage(ctx *fasthttp.RequestCtx, path string) {
	digits := secret.RandBytes(captchaWordSize, bytesPool)
	sion.Set(internal.SessionCaptchaIdKey+"."+path, toString(digits))
	sion.Set(internal.SessionCaptchaUnixKey+"."+path, time.Now().Unix())

	ctx.Response.Header.Set("Cache-control", "no-store")
	ctx.Response.Header.Set("Content-type", "image/png")

	image := captcha.NewImage(string(sion), digits, captchaWidth, captchaHeight)
	_, err := image.WriteTo(output.NewCompressionWriter(ctx))
	if err != nil {
		output.Error(ctx, err)
		return
	}
}

func (sion Session) CaptchaGenerateAudio(ctx *fasthttp.RequestCtx, path string) {
	digits := secret.RandBytes(captchaWordSize, bytesPool)
	sion.Set(internal.SessionCaptchaIdKey+"."+path, toString(digits))
	sion.Set(internal.SessionCaptchaUnixKey+"."+path, time.Now().Unix())

	ctx.Response.Header.Set("Cache-control", "no-store")
	ctx.Response.Header.Set("Content-type", "audio/wav")

	audio := captcha.NewAudio(string(sion), digits, captchaAudioLang)
	_, err := audio.WriteTo(output.NewCompressionWriter(ctx))
	if err != nil {
		output.Error(ctx, err)
		return
	}
}

func (sion Session) CaptchaVerify(ctx *fasthttp.RequestCtx, path string) (ok bool) {
	if skipVerify {
		return true
	}

	unixKey := internal.SessionCaptchaUnixKey + "." + path
	idKey := internal.SessionCaptchaIdKey + "." + path

	defer func() {
		sion.Del(unixKey, idKey) // del captcha anyway
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
	if !sion.Get(idKey, &code) {
		return
	}

	var unix int64
	if !sion.Get(unixKey, &unix) || time.Now().Unix()-unix > captchaMaxage {
		return
	}

	if code != gotils.B2S(v) {
		return
	}
	ok = true
	return
}
