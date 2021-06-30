package utils

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"strings"
	"sync"
)

type RedisConfig struct {
	Mode     string   `toml:"mode"`
	User     string   `toml:"user"`
	Password string   `toml:"password"`
	DB       int      `toml:"db"`
	Addrs    []string `toml:"addrs"`

	internal struct {
		once sync.Once
		cli  redis.Cmdable
	} `toml:"-"`
}

func (cfg *RedisConfig) Cli() redis.Cmdable {
	cfg.internal.once.Do(func() { cfg.buildRedisCli() })
	return cfg.internal.cli
}

func (cfg *RedisConfig) buildRedisCli() (cli redis.Cmdable) {
	defer func() {
		cfg.internal.cli = cli
	}()

	switch strings.ToLower(cfg.Mode) {
	case "singleton":
		opt := &redis.Options{Username: cfg.User, Password: cfg.Password, DB: cfg.DB}
		if len(cfg.Addrs) > 0 {
			opt.Addr = cfg.Addrs[0]
		}
		cli = redis.NewClient(opt)
		return
	case "cluster":
		opt := &redis.ClusterOptions{Username: cfg.User, Password: cfg.Password, Addrs: cfg.Addrs}
		cli = redis.NewClusterClient(opt)
		return
	case "ring":
		opt := &redis.RingOptions{Username: cfg.User, Password: cfg.Password, DB: cfg.DB}
		m := map[string]string{}
		for idx, a := range cfg.Addrs {
			m[fmt.Sprintf("n%d", idx)] = a
		}
		opt.Addrs = m
		cli = redis.NewRing(opt)
		return
	default:
		panic(fmt.Errorf("sha.utils: unknown redis mode `%s`, {singleton,cluster,ring}", cfg.Mode))
	}
}
