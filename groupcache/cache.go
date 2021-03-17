package groupcache

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"reflect"
	"time"
)

type NamedArgs interface {
	Name() string
}

type DataLoader func(ctx context.Context, args NamedArgs) (ret interface{}, err error)

type Convertor interface {
	Encode(v interface{}) ([]byte, error)
	Decode(v []byte, dist interface{}) error
}

type _StdJsonConvertor struct{}

func (_StdJsonConvertor) Encode(v interface{}) ([]byte, error) { return jsonx.Marshal(v) }

func (_StdJsonConvertor) Decode(v []byte, dist interface{}) error { return jsonx.Unmarshal(v, dist) }

var StdJsonConvertor = _StdJsonConvertor{}

type Storage interface {
	Set(ctx context.Context, k string, v []byte, expires time.Duration)
	Get(ctx context.Context, k string) ([]byte, bool)
	Del(ctx context.Context, keys ...string)
}

type Expires struct {
	Default utils.TomlDuration `json:"default" toml:"default"`
	Missing utils.TomlDuration `json:"missing" toml:"missing"`
}

var defaultExpires = Expires{
	Default: utils.TomlDuration{Duration: time.Hour * 5},
	Missing: utils.TomlDuration{Duration: time.Minute * 2},
}

type Group struct {
	prefix  string
	name    string
	expires Expires
	sf      *internal.SingleflightGroup
	loads   map[string]DataLoader
	storage Storage
	conv    Convertor
}

func New(name string, expires *Expires, maxWait int32) *Group {
	ret := Simple()
	ret.Configure(name, expires, maxWait)
	return ret
}

func Simple() *Group { return &Group{loads: map[string]DataLoader{}} }

func (g *Group) Configure(name string, expires *Expires, maxWait int32) *Group {
	g.name = name
	if expires == nil {
		expires = &defaultExpires
	}
	g.expires = *expires
	g.sf = internal.NewSingleflightGroup(maxWait)
	return g
}

func (g *Group) SetStoragePrefix(prefix string) *Group {
	g.prefix = prefix
	return g
}

func (g *Group) SetStorage(s Storage) *Group {
	g.storage = s
	return g
}

func (g *Group) SetRedisStorage(r redis.Cmdable) *Group {
	return g.SetStorage(RedisStorage(r))
}

func (g *Group) SetConvertor(c Convertor) *Group {
	g.conv = c
	return g
}

func (g *Group) Register(name string, loader DataLoader) *Group {
	g.loads[name] = loader
	return g
}

var emptyVal = []byte("")

var ErrEmpty = errors.New("sha.groupcache: empty")
var ErrBadCacheValue = errors.New("sha.groupcache: bad cache")
var ErrRetryAfter = internal.ErrRetryAfter

func (g *Group) MakeKey(loaderName, argsName string) string {
	if len(g.prefix) < 1 {
		return fmt.Sprintf("groupcache:%s:%s:%s", g.name, loaderName, argsName)
	}
	return fmt.Sprintf("%s:%s:%s:%s", g.prefix, g.name, loaderName, argsName)
}

// dist must be a pointer
func (g *Group) DoWithExpires(ctx context.Context, loaderName string, dist interface{}, args NamedArgs, expires *Expires) error {
	if expires == nil {
		expires = &g.expires
	}

	conv := g.conv
	if conv == nil {
		conv = StdJsonConvertor
	}

	key := g.MakeKey(loaderName, args.Name())
	v, found := g.storage.Get(ctx, key)
	if found {
		if len(v) == 0 { // empty cache
			return ErrEmpty
		}

		err := conv.Decode(v, dist)
		if err != nil { // bad groupcache
			g.storage.Del(ctx, key)
			return ErrBadCacheValue
		} else {
			return nil
		}
	}

	ret, e := g.sf.Do(
		args.Name(),
		func() (interface{}, error) {
			fn := g.loads[loaderName]
			if fn == nil {
				panic(fmt.Errorf("sha.groupcache: name `%s` is not exists in group `%s`", loaderName, g.name))
			}

			_v, e := fn(ctx, args)
			if e != nil {
				return nil, e
			}

			if _v == nil {
				g.storage.Set(ctx, key, emptyVal, expires.Missing.Duration)
				return nil, ErrEmpty
			}

			_b, e := conv.Encode(_v)
			if e != nil {
				panic(e)
			}
			g.storage.Set(ctx, key, _b, expires.Default.Duration)
			return _v, nil
		},
	)

	if e == nil {
		_copy(dist, ret)
	}
	return e
}

func (g *Group) Do(ctx context.Context, loaderName string, dist interface{}, args NamedArgs) error {
	return g.DoWithExpires(ctx, loaderName, dist, args, nil)
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
