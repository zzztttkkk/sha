package ini

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var storage = make(map[string][]byte)
var lock = &sync.Mutex{}
var (
	isDebug   bool
	isRelease bool
	isTest    bool
)

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

func Get(key string) (v []byte, e bool) {
	v, e = storage[key]
	return
}

func GetMust(key string) []byte {
	v, ok := Get(key)
	if ok {
		return v
	}
	panic(fmt.Errorf("snow.ini: `%s` is not found", key))
}

func GetIntMust(key string) int64 {
	i, err := strconv.ParseInt(string(GetMust(key)), 10, 64)
	if err != nil {
		panic(fmt.Errorf("snow.ini: `%s` is not an int", key))
	}
	return i
}

func GetOr(key string, defaultV string) string {
	v, ok := Get(key)
	if !ok {
		return defaultV
	}
	return string(v)
}

func GetIntOr(key string, defaultV int64) int64 {
	v, ok := Get(key)
	if !ok {
		return defaultV
	}

	i, e := strconv.ParseInt(string(v), 10, 64)
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
	return strings.ToLower(string(v)) == "true"
}

func IsRelease() bool {
	return isRelease
}

func IsDebug() bool {
	return isDebug
}

func IsTest() bool {
	return isTest
}

const (
	modeRelease = "release"
	modeDebug   = "debug"
	modeTest    = "test"
)
