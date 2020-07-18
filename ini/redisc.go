package ini

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis/v7"
)

var redisUnknownModeError = errors.New("snow.redisc: unknown redis mode,[singleton,ring,cluster]")
var redisModes = map[string]bool{"singleton": true, "ring": true, "cluster": true}
var redisInitError = errors.New("snow.redisc: init error")

func makeRedisKey(n string) string {
	return fmt.Sprintf("redis.%s", n)
}

func (conf *Config) initRedisClient() {
	mode := string(conf.GetMust(makeRedisKey("mode")))
	if _, ok := redisModes[mode]; !ok {
		panic(redisUnknownModeError)
	}

	var nodesCount = 0
	var err error
	if mode == "singleton" {
		nodesCount = 1
	} else {
		nodesCount = int(conf.GetIntMust(makeRedisKey("count")))
	}

	urls := make([]string, 0)
	for i := 0; i < nodesCount; i++ {
		urls = append(urls, string(conf.GetMust(makeRedisKey(fmt.Sprintf("node%d.url", i)))))
	}

	var option *redis.Options

	if mode == "singleton" {
		option, err = redis.ParseURL(urls[0])
		if err != nil {
			panic(err)
		}
		client := redis.NewClient(option)
		conf.rcmd = client
		return
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
		client := redis.NewRing(
			&redis.RingOptions{
				Addrs:     nodeMap,
				Passwords: passwordMaps,
			},
		)
		conf.rcmd = client
		return
	}

	if mode == "cluster" {
		client := redis.NewClusterClient(
			&redis.ClusterOptions{
				Addrs:    nodeLst,
				Password: passwordLst[0],
			},
		)
		conf.rcmd = client
		return
	}
	panic(redisInitError)
}
