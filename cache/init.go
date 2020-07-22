package cache

import "github.com/go-redis/redis/v7"

var c redis.Cmdable

func Init(rc redis.Cmdable) {
	c = rc
}
