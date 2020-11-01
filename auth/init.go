package auth

import "github.com/zzztttkkk/suna/internal"

func init() {
	internal.Dig.Append(func(aor Authenticator) { authenticatorV = aor })
}
