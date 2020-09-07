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
var crsfForm = ""

func (sion Session) CrsfGenerate(ctx *fasthttp.RequestCtx) string {
	key := gotils.B2S(secret.RandBytes(crsfKeyLength, nil))
	sion.Set(internal.SessionCrsfValueKey, key)
	sion.Set(internal.SessionCrsfUnixKey, time.Now().Unix())
	return key
}

func (sion Session) CrsfVerify(ctx *fasthttp.RequestCtx) bool {
	if skipVerify {
		return true
	}

	var formV = ctx.FormValue(crsfForm)
	if len(formV) != crsfKeyLength {
		return false
	}

	var key string
	if !sion.Get(internal.SessionCrsfValueKey, &key) || len(key) != crsfKeyLength {
		return false
	}
	var unix int64
	if !sion.Get(internal.SessionCrsfUnixKey, &unix) || time.Now().Unix()-unix >= crsfMaxAge {
		return false
	}
	return gotils.B2S(formV) == key
}
