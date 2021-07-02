package session

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/zzztttkkk/sha/jsonx"
	"time"
)

type Backend interface {
	NewSession(ctx context.Context, session string, duration time.Duration) error
	ExpireSession(ctx context.Context, session string, duration time.Duration) error
	TTLSession(ctx context.Context, session string) time.Duration
	ClearSession(ctx context.Context, session string) error

	SessionSet(ctx context.Context, session string, key string, val []byte) error
	SessionGet(ctx context.Context, session string, key string) ([]byte, error)
	SessionDel(ctx context.Context, session string, keys ...string) error
	SessionIncr(ctx context.Context, session string, key string, val int64) (int64, error)
	SessionGetAll(ctx context.Context, session string) map[string]string

	Set(ctx context.Context, key string, val interface{}, duration time.Duration) error
	Get(ctx context.Context, key string, dist interface{}) bool
	Del(ctx context.Context, keys ...string) error
}

type _RedisSessionBackend struct {
	cmd        redis.Cmdable
	updateSha1 string
	clearSha1  string
	newSha1    string
}

func (rsb *_RedisSessionBackend) TTLSession(ctx context.Context, session string) time.Duration {
	v, _ := rsb.cmd.TTL(ctx, session).Result()
	return v
}

var updateScript = `
redis.call('hset', KEYS[1], KEYS[2], ARGV[1]);
redis.call('expire', KEYS[1], ARGV[2]);
`

var clearScript = `
redis.call('del', KEYS[1]);
redis.call('hset', KEYS[1], '.created', ARGV[1]);
`

var newScript = `
redis.call('hset', KEYS[1], '.created', ARGV[1]);
redis.call('expire', KEYS[1], ARGV[2]);
`

func (rsb *_RedisSessionBackend) Set(ctx context.Context, key string, val interface{}, duration time.Duration) error {
	v, e := jsonx.Marshal(val)
	if e != nil {
		return e
	}
	return rsb.cmd.Set(ctx, key, v, duration).Err()
}

func (rsb *_RedisSessionBackend) Get(ctx context.Context, key string, dist interface{}) bool {
	v, e := rsb.cmd.Get(ctx, key).Bytes()
	if e != nil {
		return false
	}
	return jsonx.Unmarshal(v, dist) == nil
}

func (rsb *_RedisSessionBackend) Del(ctx context.Context, keys ...string) error {
	return rsb.cmd.Del(ctx, keys...).Err()
}

func (rsb *_RedisSessionBackend) ClearSession(ctx context.Context, session string) error {
	return rsb.cmd.EvalSha(ctx, rsb.clearSha1, []string{session}, time.Now().Unix()).Err()
}

func (rsb *_RedisSessionBackend) SessionGetAll(ctx context.Context, session string) map[string]string {
	v, _ := rsb.cmd.HGetAll(ctx, session).Result()
	return v
}

func (rsb *_RedisSessionBackend) NewSession(ctx context.Context, session string, duration time.Duration) error {
	return rsb.cmd.EvalSha(ctx, rsb.newSha1, []string{session}, time.Now().Unix(), duration).Err()
}

func (rsb *_RedisSessionBackend) ExpireSession(ctx context.Context, session string, duration time.Duration) error {
	return rsb.cmd.Expire(ctx, session, duration).Err()
}

func (rsb *_RedisSessionBackend) SessionSet(ctx context.Context, session string, key string, val []byte) error {
	return rsb.cmd.EvalSha(
		ctx, rsb.updateSha1,
		[]string{session, key},
		val, opts.MaxAge.Duration/time.Second,
	).Err()
}

func (rsb *_RedisSessionBackend) SessionGet(ctx context.Context, session string, key string) ([]byte, error) {
	return rsb.cmd.HGet(ctx, session, key).Bytes()
}

func (rsb *_RedisSessionBackend) SessionDel(ctx context.Context, session string, keys ...string) error {
	return rsb.cmd.HDel(ctx, session, keys...).Err()
}

func (rsb *_RedisSessionBackend) SessionIncr(ctx context.Context, session string, key string, val int64) (int64, error) {
	return rsb.cmd.HIncrBy(ctx, session, key, val).Result()
}

var _ Backend = (*_RedisSessionBackend)(nil)

func NewRedisBackend(cmd redis.Cmdable) Backend {
	v := &_RedisSessionBackend{cmd: cmd}
	sha1, err := cmd.ScriptLoad(context.Background(), updateScript).Result()
	if err != nil {
		panic(err)
	}
	v.updateSha1 = sha1
	sha1, err = cmd.ScriptLoad(context.Background(), clearScript).Result()
	if err != nil {
		panic(err)
	}
	v.clearSha1 = sha1
	sha1, err = cmd.ScriptLoad(context.Background(), newScript).Result()
	if err != nil {
		panic(err)
	}
	v.newSha1 = sha1
	return v
}
