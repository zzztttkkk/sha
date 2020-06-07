package ini

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var storage = make(map[string]string)
var lock = &sync.Mutex{}

func Load(filename string) {
	lock.Lock()
	defer lock.Unlock()

	for k, v := range parseIniFile(filename) {
		storage[k] = v
	}
}

func Print() {
	var ks []string

	for k := range storage {
		ks = append(ks, k)
	}
	sort.StringSlice(ks).Sort()

	for _, k := range ks {
		log.Printf("%s: %s\n", k, storage[k])
	}
}

func Get(key string) (v string, e bool) {
	v, e = storage[key]
	return
}

func MustGet(key string) string {
	v, ok := Get(key)
	if ok {
		return v
	}
	panic(fmt.Errorf("snow.ini: `%s` is not found", key))
}

func MustGetInt(key string) int64 {
	i, err := strconv.ParseInt(MustGet(key), 10, 64)
	if err != nil {
		panic(fmt.Errorf("snow.ini: `%s` is not an int", key))
	}
	return i
}

func GetOr(key string, defaultV string) string {
	v, ok := Get(key)
	if !ok {
		v = defaultV
	}
	return v
}

func GetOrInt(key string, defaultV int64) int64 {
	v, ok := Get(key)
	if !ok {
		return defaultV
	}

	i, e := strconv.ParseInt(v, 10, 64)
	if e != nil {
		return defaultV
	}
	return i
}

func GetBool(key string) bool {
	v, ok := Get(key)
	if !ok {
		return false
	}
	return strings.ToLower(v) == "true"
}

var (
	mode int
)

func IsRelease() bool {
	return mode == 0
}

func IsDebug() bool {
	return mode == 1
}

func IsTest() bool {
	return mode == 2
}

const (
	ModeRelease = "release"
	ModeDebug   = "debug"
	ModeTest    = "test"
)
