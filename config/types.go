package config

import (
	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/utils"
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
		Maxage  utils.TomlDuration
		Captcha struct {
			Maxage      int
			Form        string
			Words       int
			Width       int
			Height      int
			SkipInDebug bool `toml:"skip-in-debug"`
		}
	}

	Errors struct {
		MaxDepth int
	}

	Sql struct {
		Driver          string
		Leader          string
		Followers       []string
		MaxOpen         int                `toml:"max-open"`
		MaxLifetime     utils.TomlDuration `toml:"max-lifetime"`
		EnumCacheMaxage utils.TomlDuration `toml:"enum-cache-maxage"`
		Log             bool

		l   *sqlx.DB `toml:"-"`
		nfs bool
		fs  []*sqlx.DB `toml:"-"`
	}

	Rbac struct {
		TablenamePrefix string `toml:"tablename-prefix"`
	}

	Redis struct {
		Mode  string
		Nodes []string

		c redis.Cmdable `toml:"-"`
	}

	isDebug   bool
	isRelease bool
	isTest    bool
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
	defaultV.Session.Captcha.Width = 640
	defaultV.Session.Captcha.Words = 6
	defaultV.Session.Captcha.Maxage = 300
	defaultV.Rbac.TablenamePrefix = "rbac_"
}

func GetDefault() *Suna { return &defaultV }

func (t *Suna) Done() {
	t.Sql.Driver = strings.ToLower(t.Sql.Driver)
	t.Redis.Mode = strings.ToLower(t.Redis.Mode)

	switch strings.ToLower(t.Env) {
	case "debug":
		t.isDebug = true
	case "release":
		t.isRelease = true
	case "test":
		t.isTest = true
	default:
		t.isRelease = true
	}
}

func (t *Suna) IsDebug() bool { return t.isDebug }

func (t *Suna) IsRelease() bool { return t.isRelease }

func (t *Suna) IsTest() bool { return t.isTest }
