package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v7"
)

// ErrRedisUnknownMode unknown redis mode error
var ErrRedisUnknownMode = errors.New("suna.config: unknown redis mode,[singleton,ring,cluster]")

//revive:disable:cyclomatic
func (t *Suna) _InitRedisClient() {
	if t.Internal.redisc != nil {
		return
	}

	defer func() {
		if t.Internal.redisc != nil {
			if err := t.Internal.redisc.Ping().Err(); err != nil {
				panic(err)
			}
		}
	}()

	opts := make([]*redis.Options, 0)
	for _, url := range t.Redis.Nodes {
		option, err := redis.ParseURL(url)
		if err != nil {
			panic(err)
		}
		opts = append(opts, option)
	}

	if len(opts) < 1 {
		return
	}

	switch strings.ToLower(t.Redis.Mode) {
	case "singleton":
		t.Internal.redisc = redis.NewClient(opts[0])
		return
	case "ring":
		addrs := map[string]string{}
		pwds := map[string]string{}
		for ind, opt := range opts {
			addrs[fmt.Sprintf("node.%d", ind)] = opt.Addr
			pwds[fmt.Sprintf("node.%d", ind)] = opt.Password
		}
		t.Internal.redisc = redis.NewRing(&redis.RingOptions{Addrs: addrs, Passwords: pwds})
		return
	case "cluster":
		var addrs []string
		for _, opt := range opts {
			addrs = append(addrs, opt.Addr)
		}
		t.Internal.redisc = redis.NewClusterClient(
			&redis.ClusterOptions{
				Addrs:    addrs,
				Username: opts[0].Username,
				Password: opts[0].Password,
			},
		)
		return
	default:
		panic(ErrRedisUnknownMode)
	}
}

// RedisClient get redis client
func (t *Suna) RedisClient() redis.Cmdable {
	if t.Internal.redisc == nil {
		t._InitRedisClient()
	}
	return t.Internal.redisc
}
