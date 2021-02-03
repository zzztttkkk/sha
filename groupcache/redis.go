package groupcache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type _RedisStorage struct {
	cli redis.Cmdable
}

func (rs _RedisStorage) Set(ctx context.Context, k string, v []byte, expires time.Duration) {
	if err := rs.cli.Set(ctx, k, v, expires).Err(); err != nil {
		panic(err)
	}
}

func (rs _RedisStorage) Get(ctx context.Context, k string) ([]byte, bool) {
	v, e := rs.cli.Get(ctx, k).Bytes()
	if e != nil {
		if e == redis.Nil {
			return nil, false
		}
		panic(e)
	}
	return v, true
}

func (rs _RedisStorage) Del(ctx context.Context, keys ...string) { rs.cli.Del(ctx, keys...) }

func RedisStorage(cmd redis.Cmdable) Storage { return _RedisStorage{cli: cmd} }
