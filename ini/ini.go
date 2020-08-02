package ini

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"

	"github.com/zzztttkkk/suna/utils"
)

type Ini struct {
	storage map[string][]byte
	raw     map[string]string

	isDebug   bool
	isRelease bool
	isTest    bool

	sqlL *sqlx.DB
	sqlF []*sqlx.DB

	rcmd redis.Cmdable

	files []string
}

func New() *Ini {
	return &Ini{
		storage: map[string][]byte{},
		raw:     map[string]string{},
	}
}

func (conf *Ini) Load(filename string) {
	conf.files = append(conf.files, filename)
	parseIniFile(conf, filename)
}

func (conf *Ini) Print() {
	var ks []string
	glog := utils.AcquireGroupLogger(fmt.Sprintf("Config: [%s]", strings.Join(conf.files, "; ")))
	defer utils.ReleaseGroupLogger(glog)

	for k := range conf.raw {
		ks = append(ks, k)
	}
	sort.StringSlice(ks).Sort()

	for _, k := range ks {
		glog.Println(fmt.Sprintf("%s: %s", k, conf.raw[k]))
	}
}

func (conf *Ini) Get(key string) (v []byte, e bool) {
	v, e = conf.storage[key]
	return
}

func (conf *Ini) GetMust(key string) []byte {
	v, ok := conf.Get(key)
	if ok {
		return v
	}
	panic(fmt.Errorf("suna.ini: `%s` is not found", key))
}

func (conf *Ini) GetIntMust(key string) int64 {
	i, err := strconv.ParseInt(string(conf.GetMust(key)), 10, 64)
	if err != nil {
		panic(fmt.Errorf("suna.ini: `%s` is not an int", key))
	}
	return i
}

func (conf *Ini) GetOr(key string, defaultV string) string {
	v, ok := conf.Get(key)
	if !ok {
		return defaultV
	}
	return string(v)
}

func (conf *Ini) GetIntOr(key string, defaultV int64) int64 {
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

func (conf *Ini) GetBool(key string) bool {
	v, ok := conf.Get(key)
	if !ok {
		return false
	}
	return strings.ToLower(string(v)) == "true"
}

func (conf *Ini) IsRelease() bool {
	return conf.isRelease
}

func (conf *Ini) IsDebug() bool {
	return conf.isDebug
}

func (conf *Ini) IsTest() bool {
	return conf.isTest
}

const (
	modeRelease = "release"
	modeDebug   = "debug"
	modeTest    = "test"
)

func (conf *Ini) Done() {
	switch strings.ToLower(conf.GetOr("app.mode", "debug")) {
	case modeRelease:
		conf.isRelease = true
	case modeDebug:
		conf.isDebug = true
	case modeTest:
		conf.isTest = true
	default:
		panic(fmt.Errorf("suna.ini: unknown `app.mode`"))
	}
}
