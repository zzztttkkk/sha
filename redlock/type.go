package redlock

import (
	"context"
	"errors"
	"time"
)

type RedLock struct {
	key string
	du  time.Duration
	cfn func()
}

func New(key string, expire time.Duration) *RedLock {
	return &RedLock{
		key: key,
		du:  expire,
		cfn: nil,
	}
}

var ErrKeyExists = errors.New("suna.redlock: the key is exists")
var ErrTimeout = errors.New("suna.redlock: acquire timeout")

func (lock *RedLock) Acquire(ctx context.Context) (context.Context, error) {
	begin := time.Now()
	if !redisc.SetNX(lock.key, 1, lock.du).Val() {
		return nil, ErrKeyExists
	}

	du := lock.du - (time.Now().Sub(begin))
	if du <= 0 {
		return nil, ErrTimeout
	}
	if du > time.Second {
		du = du - time.Millisecond*30
	}
	nctx, cfn := context.WithTimeout(ctx, du)
	lock.cfn = cfn
	return nctx, nil
}

func (lock *RedLock) Release(alsoReleaseRedisKey bool) {
	if alsoReleaseRedisKey {
		redisc.Del(lock.key)
	}
	lock.cfn()
}
