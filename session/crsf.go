package session

import (
	"time"

	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/secret"
)

var crsfKeyLength = 16
var crsfMaxAge = int64(300)
var crsfForm = ""

func (sion Session) CrsfGenerate(path string) string {
	key := gotils.B2S(secret.RandBytes(crsfKeyLength, nil))
	sion.Set(internal.SessionCrsfValueKey+"."+path, key)
	sion.Set(internal.SessionCrsfUnixKey+"."+path, time.Now().Unix())
	return key
}

func (sion Session) CrsfVerify(ctx *fasthttp.RequestCtx, path string) bool {
	if skipVerify {
		return true
	}

	var formV = ctx.FormValue(crsfForm)
	if len(formV) != crsfKeyLength {
		return false
	}

	var key string
	if !sion.Get(internal.SessionCrsfValueKey+"."+path, &key) || len(key) != crsfKeyLength {
		return false
	}
	var unix int64
	if !sion.Get(internal.SessionCrsfUnixKey+"."+path, &unix) || time.Now().Unix()-unix >= crsfMaxAge {
		return false
	}
	return gotils.B2S(formV) == key
}
