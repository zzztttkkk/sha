package session

import (
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/secret"
	"time"
)

var crsfKeyLength = 16
var crsfMaxAge = int64(300)

func (sion Session) CrsfGenerate(ctx *fasthttp.RequestCtx) string {
	key := gotils.B2S(secret.RandBytes(crsfKeyLength, nil))
	sion.Set(internal.SessionCrsfValueKey, key)
	sion.Set(internal.SessionCrsfUnixKey, time.Now().Unix())
	return key
}

func (sion Session) CrsfVerify(ctx *fasthttp.RequestCtx) (ok bool) {
	if captchaSkipVerify {
		return true
	}

	var key string
	if !sion.Get(internal.SessionCrsfValueKey, &key) || len(key) != crsfKeyLength {
		return
	}
	var unix int64
	if !sion.Get(internal.SessionCrsfUnixKey, &unix) || time.Now().Unix()-unix >= crsfMaxAge {
		return
	}

	return false
}
