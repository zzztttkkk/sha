package config

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
)

var redisUnknownModeError = errors.New("suna.config: unknown redis mode,[singleton,ring]")

func (t *Type) RedisClient() redis.Cmdable {
	if t.Redis.c != nil {
		return t.Redis.c
	}

	opts := make([]*redis.Options, 0)
	for _, url := range t.Redis.Nodes {
		option, err := redis.ParseURL(url)
		if err != nil {
			panic(err)
		}
		opts = append(opts, option)
	}

	if len(opts) < 1 {
		return nil
	}

	switch t.Redis.Mode {
	case "singleton":
		t.Redis.c = redis.NewClient(opts[0])
		return t.Redis.c
	case "ring":
		addrs := map[string]string{}
		pwds := map[string]string{}
		for ind, opt := range opts {
			addrs[fmt.Sprintf("node.%d", ind)] = opt.Addr
			pwds[fmt.Sprintf("node.%d", ind)] = opt.Password
		}
		t.Redis.c = redis.NewRing(&redis.RingOptions{Addrs: addrs, Passwords: pwds})
		return t.Redis.c
	default:
		panic(redisUnknownModeError)
	}
}
