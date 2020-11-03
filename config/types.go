package config

import (
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
)

type Suna struct {
	Env string

	Http struct {
		Address  string
		Hostname string
		TLS      struct {
			Cert string
			Key  string
		}
	}

	Captcha struct {
		Fonts []string `toml:"fonts"` // <fontName>:<fontPath>[:fontSize]  A:/fonts/a.ttf:16 B:/fonts/b.rrf
	}

	Secret struct {
		Key           string
		HashAlgorithm string `toml:"hash-algorithm"`
	}

	Cache struct {
		Lru struct {
			UserSize    int `toml:"user-size"`
			ContentSize int `toml:"content-size"`
		}
	}

	Session struct {
		HeaderName     string   `toml:"header-name"`
		CookieName     string   `toml:"cookie-name"`
		RedisKeyPrefix string   `toml:"redis-key-prefix"`
		MaxAge         Duration `toml:"max-age"`
	}

	Output struct {
		ErrorMaxDepth     int    `toml:"error-max-depth"`
		JsonPCallbackForm string `toml:"jsonp-callback-form"`
		Compress          bool   `toml:"compress"`
	}

	Sql struct {
		Driver          string
		Leader          string
		Followers       []string
		MaxOpen         int      `toml:"max-open"`
		MaxLifetime     Duration `toml:"max-lifetime"`
		EnumCacheMaxAge Duration `toml:"enum-cache-maxage"`
		Logging         bool
	}

	Rbac struct {
		TablenamePrefix string `toml:"tablename-prefix"`
	}

	Redis struct {
		Mode  string
		Nodes []string
	}

	internal struct {
		isDebug   bool
		isRelease bool
		isTest    bool

		redisc redis.Cmdable

		sqlLeader    *sqlx.DB
		sqlFollowers []*sqlx.DB
	} `toml:"-"`
}

func makeDefault() *Suna {
	defaultV := &Suna{}

	defaultV.Env = "debug"
	defaultV.Http.Address = "127.0.0.1:8080"
	defaultV.Secret.HashAlgorithm = "sha256-512"
	defaultV.Cache.Lru.ContentSize = 2000
	defaultV.Cache.Lru.UserSize = 1000
	defaultV.Output.ErrorMaxDepth = 20
	defaultV.Session.CookieName = "session"
	defaultV.Session.HeaderName = "Session"
	defaultV.Session.MaxAge.Duration = time.Minute * 30
	defaultV.Session.RedisKeyPrefix = "session:"
	defaultV.Rbac.TablenamePrefix = "rbac_"

	return defaultV
}

func Default() *Suna { return makeDefault() }

func (t *Suna) Done() {
	t.Redis.Mode = strings.ToLower(t.Redis.Mode)

	switch strings.ToLower(t.Env) {
	case "debug":
		t.internal.isDebug = true
	case "release":
		t.internal.isRelease = true
	case "test":
		t.internal.isTest = true
	default:
		t.internal.isDebug = true
	}

	t._InitRedisClient()
}

func (t *Suna) IsDebug() bool { return t.internal.isDebug }

func (t *Suna) IsRelease() bool { return t.internal.isRelease }

func (t *Suna) IsTest() bool { return t.internal.isTest }
