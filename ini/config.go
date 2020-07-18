package ini

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"

	"github.com/zzztttkkk/snow/utils"
)

type Config struct {
	storage map[string][]byte
	raw     map[string]string

	isDebug   bool
	isRelease bool
	isTest    bool

	sqlL *sqlx.DB
	sqlF []*sqlx.DB

	rcmd redis.Cmdable
}

func New() *Config {
	return &Config{
		storage: map[string][]byte{},
		raw:     map[string]string{},
	}
}

func (conf *Config) Load(filename string) {
	parseIniFile(conf, filename)
}

func (conf *Config) Print() {
	var ks []string
	glog := utils.AcquireGroupLogger("Config")
	defer utils.ReleaseGroupLogger(glog)

	for k := range conf.raw {
		ks = append(ks, k)
	}
	sort.StringSlice(ks).Sort()

	for _, k := range ks {
		glog.Println(fmt.Sprintf("%s: %s", k, conf.raw[k]))
	}
}

func (conf *Config) Get(key string) (v []byte, e bool) {
	v, e = conf.storage[key]
	return
}

func (conf *Config) GetMust(key string) []byte {
	v, ok := conf.Get(key)
	if ok {
		return v
	}
	panic(fmt.Errorf("snow.ini: `%s` is not found", key))
}

func (conf *Config) GetIntMust(key string) int64 {
	i, err := strconv.ParseInt(string(conf.GetMust(key)), 10, 64)
	if err != nil {
		panic(fmt.Errorf("snow.ini: `%s` is not an int", key))
	}
	return i
}

func (conf *Config) GetOr(key string, defaultV string) string {
	v, ok := conf.Get(key)
	if !ok {
		return defaultV
	}
	return string(v)
}

func (conf *Config) GetIntOr(key string, defaultV int64) int64 {
	v, ok := conf.Get(key)
	if !ok {
		return defaultV
	}

	i, e := strconv.ParseInt(string(v), 10, 64)
	if e != nil {
		return defaultV
	}
	return i
}

func (conf *Config) GetBool(key string) bool {
	v, ok := conf.Get(key)
	if !ok {
		return false
	}
	return strings.ToLower(string(v)) == "true"
}

func (conf *Config) IsRelease() bool {
	return conf.isRelease
}

func (conf *Config) IsDebug() bool {
	return conf.isDebug
}

func (conf *Config) IsTest() bool {
	return conf.isTest
}

const (
	modeRelease = "release"
	modeDebug   = "debug"
	modeTest    = "test"
)

func (conf *Config) RedisClient() redis.Cmdable {
	return conf.rcmd
}

func (conf *Config) SqlLeader() *sqlx.DB {
	return conf.sqlL
}

func (conf *Config) SqlFollower() *sqlx.DB {
	if len(conf.sqlF) < 1 {
		return nil
	}
	rand.Seed(time.Now().UnixNano())
	return conf.sqlF[rand.Int()%len(conf.sqlF)]
}

func (conf *Config) Done() {
	switch string(conf.GetMust("app.mode")) {
	case modeRelease:
		conf.isRelease = true
	case modeDebug:
		conf.isDebug = true
	case modeTest:
		conf.isTest = true
	default:
		panic(fmt.Errorf("snow.ini: unknown `app.mode`"))
	}

	conf.initRedisClient()
	conf.initSqlClient()
}
