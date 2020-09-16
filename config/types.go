package config

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
)

type Suna struct {
	Env           string
	TimeFormatter string `toml:"time-formatter"`

	Http struct {
		Address  string
		Hostname string
		TLS      struct {
			Cert string
			Key  string
		}
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
		Header            string
		Cookie            string
		Prefix            string
		Maxage            Duration
		SkipVerifyInDebug bool `toml:"skip-verify-in-debug"`

		Captcha struct {
			Maxage        int
			Form          string
			TokenSize     int `toml:"token-size"`
			ImageWidth    int
			ImageHeight   int
			AudioLanguage string `toml:"audio-lang"`
		}

		Crsf struct {
			Maxage    int
			Form      string
			TokenSize int `toml:"token-size"`
		}
	}

	Output struct {
		ErrorMaxDepth     int    `toml:"error-max-depth"`
		JsonPCallbackForm string `toml:"jsonp-callback-form"`
	}

	Sql struct {
		Driver          string
		Leader          string
		Followers       []string
		MaxOpen         int      `toml:"max-open"`
		MaxLifetime     Duration `toml:"max-lifetime"`
		EnumCacheMaxage Duration `toml:"enum-cache-maxage"`
		Logging         bool
	}

	Rbac struct {
		TablenamePrefix string `toml:"tablename-prefix"`
	}

	Redis struct {
		Mode  string
		Nodes []string
	}

	Json struct {
		Marshal   func(v interface{}) ([]byte, error)
		Unmarshal func([]byte, interface{}) error
	} `toml:"-"`

	Internal struct {
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
	defaultV.Session.Cookie = "sck"
	defaultV.Session.Header = "Suna-Session"
	defaultV.Session.Maxage.Duration = time.Minute * 30
	defaultV.Session.Prefix = "session"
	defaultV.Session.Captcha.Form = "captcha"
	defaultV.Session.Captcha.ImageHeight = 120
	defaultV.Session.Captcha.ImageWidth = 300
	defaultV.Session.Captcha.TokenSize = 6
	defaultV.Session.Captcha.Maxage = 300
	defaultV.Session.Captcha.AudioLanguage = "zh"
	defaultV.Rbac.TablenamePrefix = "rbac_"
	defaultV.Json.Unmarshal = json.Unmarshal
	defaultV.Json.Marshal = json.Marshal

	return defaultV
}

func Default() *Suna { return makeDefault() }

func (t *Suna) Done() {
	t.Sql.Driver = strings.ToLower(t.Sql.Driver)
	t.Redis.Mode = strings.ToLower(t.Redis.Mode)

	switch strings.ToLower(t.Env) {
	case "debug":
		t.Internal.isDebug = true
	case "release":
		t.Internal.isRelease = true
	case "test":
		t.Internal.isTest = true
	default:
		t.Internal.isDebug = true
	}

	t._InitRedisClient()
}

func (t *Suna) IsDebug() bool { return t.Internal.isDebug }

func (t *Suna) IsRelease() bool { return t.Internal.isRelease }

func (t *Suna) IsTest() bool { return t.Internal.isTest }
