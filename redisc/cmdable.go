package redisc

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/zzztttkkk/snow/ini"
	"strconv"
)

var cmdable redis.Cmdable

var redisUnknownModeError = errors.New("snow.redisc: unknown redis mode,[singleton,ring,cluster]")
var redisModes = map[string]bool{"singleton": true, "ring": true, "cluster": true}
var redisInitError = errors.New("snow.redisc: init error")

func makeKey(n string) string {
	return fmt.Sprintf("redis.%s", n)
}

func redisInitByIni() (client redis.Cmdable) {
	mode := ini.MustGet(makeKey("mode"))
	if _, ok := redisModes[mode]; !ok {
		panic(redisUnknownModeError)
	}

	var nodesCount = 0
	var err error
	if mode == "singleton" {
		nodesCount = 1
	} else {
		nodesCount, err = strconv.Atoi(ini.MustGet(makeKey("count")))
		if err != nil {
			panic(err)
		}
	}

	urls := make([]string, 0)
	for i := 0; i < nodesCount; i++ {
		urls = append(urls, ini.MustGet(makeKey(fmt.Sprintf("node%d.url", i))))
	}

	var option *redis.Options

	if mode == "singleton" {
		option, err = redis.ParseURL(urls[0])
		if err != nil {
			panic(err)
		}
		return redis.NewClient(option)
	}

	nodeMap := make(map[string]string)
	passwordMaps := make(map[string]string)
	nodeLst := make([]string, 0)
	passwordLst := make([]string, 0)

	for _, addr := range urls {
		option, err = redis.ParseURL(addr)
		if err != nil {
			panic(err)
		}

		nodeLst = append(nodeLst, option.Addr)
		passwordLst = append(passwordLst, option.Password)
		nodeMap[option.Addr] = option.Addr
		passwordMaps[option.Addr] = option.Password
	}

	if mode == "ring" {
		return redis.NewRing(
			&redis.RingOptions{
				Addrs:     nodeMap,
				Passwords: passwordMaps,
			},
		)
	}

	if mode == "cluster" {
		return redis.NewClusterClient(
			&redis.ClusterOptions{
				Addrs:    nodeLst,
				Password: passwordLst[0],
			},
		)
	}

	panic(redisInitError)
}

func Init() {
	cmdable = redisInitByIni()
}

func Client() redis.Cmdable {
	if cmdable == nil {
		Init()
	}
	return cmdable
}
