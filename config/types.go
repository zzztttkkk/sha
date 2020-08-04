package config

import (
	"github.com/BurntSushi/toml"
	"github.com/go-redis/redis/v7"
	"github.com/imdario/mergo"
	"github.com/jmoiron/sqlx"
	"github.com/zzztttkkk/suna/auth"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type Type struct {
	Env           string
	Authenticator auth.Authenticator

	Secret struct {
		Key           string
		HashAlgorithm string
	}

	Cache struct {
		Lru struct {
			UserSize    int
			ContentSize int
		}
	}

	Session struct {
		Header  string
		Cookie  string
		Prefix  string
		MaxAge  time.Duration
		Captcha struct {
			MaxAge      int
			Form        string
			Words       int
			Width       int
			Height      int
			SkipInDebug bool
		}
	}

	Errors struct {
		MaxDepth int
	}

	Sql struct {
		Driver          string
		Leader          string
		Followers       []string
		MaxOpen         int
		MaxLifetime     time.Duration
		EnumCacheMaxAge time.Duration
		Log             bool

		l   *sqlx.DB `toml:"-"`
		nfs bool
		fs  []*sqlx.DB `toml:"-"`
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

func _New() *Type {
	t := &Type{}
	t.Env = "debug"
	t.Secret.HashAlgorithm = "sha256-512"
	t.Cache.Lru.ContentSize = 2000
	t.Cache.Lru.UserSize = 1000
	t.Errors.MaxDepth = 20
	t.Session.Cookie = "sck"
	t.Session.Header = "Suna-Session"
	t.Session.MaxAge = time.Minute * 30
	t.Session.Prefix = "session"
	t.Session.Captcha.Form = "captcha"
	t.Session.Captcha.Height = 120
	t.Session.Captcha.Width = 640
	t.Session.Captcha.Words = 6
	t.Session.Captcha.MaxAge = 300
	return t
}

func FromFile(fp string) *Type {
	f, e := os.Open(fp)
	if e != nil {
		panic(e)
	}
	defer f.Close()

	v, e := ioutil.ReadAll(f)
	if e != nil {
		panic(e)
	}
	return FromBytes(v)
}

func FromBytes(data []byte) *Type {
	conf := _New()
	if err := toml.Unmarshal(data, conf); err != nil {
		panic(err)
	}
	conf.done()
	return conf
}

func FromFiles(fps ...string) *Type {
	var t *Type
	for _, fp := range fps {
		nt := FromFile(fp)
		if t == nil {
			t = nt
		} else {
			if err := mergo.Merge(t, nt, mergo.WithOverride); err != nil {
				panic(err)
			}
		}
	}
	return t
}

func (t *Type) done() {
	t.Sql.Driver = strings.ToLower(t.Sql.Driver)
	t.Redis.Mode = strings.ToLower(t.Redis.Mode)

	switch t.Env {
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
