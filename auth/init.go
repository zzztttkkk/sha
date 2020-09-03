package auth

import "github.com/zzztttkkk/suna/internal"

func init() {
	internal.Dig.LazyInvoke(func(aor Authenticator) { _Authenticator = aor })
}
