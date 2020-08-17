package config

import (
	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/utils/toml"
	"strings"
	"time"
)

type Suna struct {
	Env           string
	TimeFormatter string `toml:"time-formatter"`

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
		Header  string
		Cookie  string
		Prefix  string
		Maxage  toml.Duration
		Captcha struct {
			Maxage        int
			Form          string
			Words         int
			Width         int
			Height        int
			AudioLanguage string `toml:"audio-lang"`
			SkipInDebug   bool   `toml:"skip-in-debug"`
		}
	}

	Errors struct {
		MaxDepth int
	}

	Sql struct {
		Driver          string
		Leader          string
		Followers       []string
		MaxOpen         int           `toml:"max-open"`
		MaxLifetime     toml.Duration `toml:"max-lifetime"`
		EnumCacheMaxage toml.Duration `toml:"enum-cache-maxage"`
		Log             bool
	}

	Rbac struct {
		TablenamePrefix string `toml:"tablename-prefix"`
	}

	Redis struct {
		Mode  string
		Nodes []string
	}

	Internal struct {
		isDebug   bool
		isRelease bool
		isTest    bool

		rediscOk bool
		redisc   redis.Cmdable

		sqlLeader        *sqlx.DB
		sqlNullFollowers bool
		sqlFollowers     []*sqlx.DB
	} `toml:"-"`
}

var defaultV = Suna{}

func init() {
	defaultV.Env = "debug"
	defaultV.Secret.HashAlgorithm = "sha256-512"
	defaultV.Cache.Lru.ContentSize = 2000
	defaultV.Cache.Lru.UserSize = 1000
	defaultV.Errors.MaxDepth = 20
	defaultV.Session.Cookie = "sck"
	defaultV.Session.Header = "Suna-Session"
	defaultV.Session.Maxage.Duration = time.Minute * 30
	defaultV.Session.Prefix = "session"
	defaultV.Session.Captcha.Form = "captcha"
	defaultV.Session.Captcha.Height = 120
	defaultV.Session.Captcha.Width = 300
	defaultV.Session.Captcha.Words = 6
	defaultV.Session.Captcha.Maxage = 300
	defaultV.Session.Captcha.AudioLanguage = "zh"
	defaultV.Rbac.TablenamePrefix = "rbac_"
}

func Default() *Suna { return &defaultV }

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
}

func (t *Suna) IsDebug() bool { return t.Internal.isDebug }

func (t *Suna) IsRelease() bool { return t.Internal.isRelease }

func (t *Suna) IsTest() bool { return t.Internal.isTest }
