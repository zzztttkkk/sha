package groupcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/utils"
	"math/rand"
	"reflect"
	"time"
)

type NamedArgs interface {
	Name() string
}

type DataLoader func(ctx context.Context, args NamedArgs) (ret interface{}, err error)

type Storage interface {
	Set(ctx context.Context, k string, v []byte, expires time.Duration)
	Get(ctx context.Context, k string) ([]byte, bool)
	Del(ctx context.Context, keys ...string)
}

type Expires struct {
	Default utils.TomlDuration `json:"default" toml:"default"`
	Missing utils.TomlDuration `json:"missing" toml:"missing"`
	Rand    int64              `json:"rand" toml:"rand"`
}

var defaultExpires = Expires{
	Default: utils.TomlDuration{Duration: time.Hour * 5},
	Missing: utils.TomlDuration{Duration: time.Minute * 2},
	Rand:    500,
}

type Group struct {
	prefix  string
	name    string
	expires Expires
	sf      *internal.SingleflightGroup
	loads   map[string]DataLoader
	storage Storage
}

func New(name string, expires *Expires, maxWait int32) *Group {
	ret := &Group{
		name:    name,
		loads:   map[string]DataLoader{},
		storage: nil,
		sf:      internal.NewSingleflightGroup(maxWait),
	}

	if expires == nil {
		expires = &defaultExpires
	}
	ret.expires = *expires
	return ret
}

func (t *Group) SetStoragePrefix(prefix string) *Group {
	t.prefix = prefix
	return t
}

func (t *Group) SetStorage(s Storage) *Group {
	t.storage = s
	return t
}

func (t *Group) SetRedisStorage(r redis.Cmdable) *Group {
	return t.SetStorage(RedisStorage(r))
}

func (t *Group) Append(name string, loader DataLoader) *Group {
	t.loads[name] = loader
	return t
}

var emptyVal = []byte("")

var ErrEmpty = errors.New("sha.groupcache: empty")
var ErrBadCacheValue = errors.New("sha.groupcache: bad cache")
var ErrRetryAfter = internal.ErrRetryAfter

func (t *Group) makeKey(loaderName, argsName string) string {
	if len(t.prefix) < 1 {
		return fmt.Sprintf("groupcache:%s:%s:%s", t.name, loaderName, argsName)
	}
	return fmt.Sprintf("%s:%s:%s:%s", t.prefix, t.name, loaderName, argsName)
}

func (t *Group) randExpires() time.Duration {
	return t.expires.Default.Duration + time.Duration(rand.Int63()%t.expires.Rand)*time.Second
}

func init() {
	utils.MathRandSeed()
}

// dist must be a pointer
func (t *Group) Do(ctx context.Context, loaderName string, dist interface{}, args NamedArgs) error {
	key := t.makeKey(loaderName, args.Name())
	v, found := t.storage.Get(ctx, key)
	if found {
		if len(v) == 0 { // empty cache
			return ErrEmpty
		}

		err := json.Unmarshal(v, dist)
		if err != nil { // bad groupcache
			t.storage.Del(ctx, key)
			return ErrBadCacheValue
		} else {
			return nil
		}
	}

	ret, e := t.sf.Do(
		args.Name(),
		func() (interface{}, error) {
			fn := t.loads[loaderName]
			if fn == nil {
				panic(fmt.Errorf("sha.groupcache: name `%s` is not exists in group `%s`", loaderName, t.name))
			}

			_v, e := fn(ctx, args)
			if e != nil {
				return nil, e
			}

			if _v == nil {
				t.storage.Set(ctx, key, emptyVal, t.expires.Missing.Duration)
				return nil, ErrEmpty
			}

			_b, e := json.Marshal(_v)
			if e != nil {
				panic(e)
			}
			t.storage.Set(ctx, key, _b, t.randExpires())
			return _v, nil
		},
	)

	if e == nil {
		_copy(dist, ret)
	}
	return e
}

// dist must be a pointer
func _copy(dist, src interface{}) {
	if dist == src {
		return
	}

	rv := reflect.ValueOf(src)
	if rv.Kind() == reflect.Ptr {
		reflect.ValueOf(dist).Elem().Set(rv.Elem())
	} else {
		reflect.ValueOf(dist).Elem().Set(rv)
	}
}
