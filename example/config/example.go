package config

import "github.com/zzztttkkk/suna/config"

type Example struct {
	config.Suna

	Auth struct {
		HeaderName string `toml:"header"`
		CookieName string `toml:"cookie"`
	}
}

var defaultV = Example{}

func init() {
	defaultV.Auth.CookieName = "alk"
	defaultV.Auth.HeaderName = "X-Auth-Token"
	defaultV.Suna = config.Default()
}

func Default() Example { return defaultV }
