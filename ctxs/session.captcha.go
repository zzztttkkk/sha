package ctxs

import (
	"github.com/dchest/captcha"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/output"
	"github.com/zzztttkkk/suna/secret"
	"github.com/zzztttkkk/suna/utils"
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

var captchaWordSize int
var captchaWidth int
var captchaHeight int

func _initCaptcha() {
	captchaWordSize = int(config.GetIntOr("session.captcha.word_size", 6))
	captchaHeight = int(config.GetIntOr("session.captcha.height", 160))
	captchaWidth = int(config.GetIntOr("session.captcha.width", int64(captchaWordSize*40)))
}

func (ss SessionStorage) CaptchaGenerate(ctx *fasthttp.RequestCtx) {
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

func (ss SessionStorage) CaptchaVerify(ctx *fasthttp.RequestCtx) (ok bool) {
	if config.IsDebug() {
		return true
	}

	defer func() {
		ss.Del(captchaIdKey, captchaUnixKey) // del captcha anyway
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
	if !ss.Get(captchaIdKey, &code) {
		return
	}

	var unix int64
	if !ss.Get(captchaUnixKey, &unix) || time.Now().Unix()-unix > 300 {
		return
	}

	if code != utils.B2s(v) {
		return
	}
	ok = true
	return
}
