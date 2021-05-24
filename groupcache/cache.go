package groupcache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/zzztttkkk/sha/internal"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"math/rand"
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

type _JSONConvertor struct{}

func (_JSONConvertor) Encode(v interface{}) ([]byte, error) { return jsonx.Marshal(v) }

func (_JSONConvertor) Decode(v []byte, dist interface{}) error { return jsonx.Unmarshal(v, dist) }

var DefaultBytesConvertor = _JSONConvertor{}

type Storage interface {
	Set(ctx context.Context, k string, v []byte, expire time.Duration)
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

type GroupOptions struct {
	Prefix     string
	Expires    Expires
	RetryLimit int
	RetrySleep time.Duration
	MaxWait    int
}

type Group struct {
	Opts    GroupOptions
	name    string
	sf      *internal.SingleflightGroup
	loads   map[string]DataLoader
	cpys    map[string]func(dist, src interface{})
	storage Storage
	conv    Convertor
}

var defaultGroupOpts = GroupOptions{
	Expires:    defaultExpires,
	RetryLimit: 0,
	RetrySleep: time.Millisecond * 50,
	MaxWait:    50,
}

func New(name string, opts *GroupOptions) *Group {
	g := &Group{loads: map[string]DataLoader{}, cpys: map[string]func(dist interface{}, src interface{}){}}
	g.name = name

	if opts == nil {
		opts = &defaultGroupOpts
	}
	g.Opts = *opts
	g.sf = internal.NewSingleflightGroup(int32(g.Opts.MaxWait))
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

func (g *Group) RegisterLoader(name string, loader DataLoader) *Group {
	g.loads[name] = loader
	return g
}

func (g *Group) RegisterCopyFunc(name string, cpy func(dist, src interface{})) *Group {
	g.cpys[name] = cpy
	return g
}

var emptyVal = []byte("")

var ErrBadCacheValue = errors.New("sha.groupcache: bad cache")
var ErrRetryAfter = internal.ErrRetryAfter
var ErrNotFound = errors.New("sha.groupcache: not found")
var errNil = errors.New("sha.groupcache: nil")

func (g *Group) MakeKey(loaderName, argsName string) string {
	if len(g.Opts.Prefix) < 1 {
		return fmt.Sprintf("groupcache:%s:%s:%s", g.name, loaderName, argsName)
	}
	return fmt.Sprintf("%s:%s:%s:%s", g.Opts.Prefix, g.name, loaderName, argsName)
}

// DoWithExpire dist must be a pointer
func (g *Group) do(ctx context.Context, loaderName string, dist interface{}, args NamedArgs) error {
	loader := g.loads[loaderName]
	if loader == nil {
		return fmt.Errorf("sha.groupcache: name `%s` is not exists in group `%s`", loaderName, g.name)
	}

	conv := g.conv
	if conv == nil {
		conv = DefaultBytesConvertor
	}

	opts := &g.Opts

	ret, e := g.sf.Do(
		args.Name(),
		func() (interface{}, error) {
			key := g.MakeKey(loaderName, args.Name())
			_bytes, found := g.storage.Get(ctx, key)

			if found {
				if len(_bytes) == 0 { // empty cache
					return nil, ErrNotFound
				}

				err := conv.Decode(_bytes, dist)
				if err != nil { // bad cached value or bad dist, but i trust the caller
					g.storage.Del(ctx, key)
					return nil, ErrBadCacheValue
				}
				return dist, nil
			}

			_val, e := loader(ctx, args)
			if e != nil {
				if e != sql.ErrNoRows && e != ErrNotFound {
					return nil, e
				}
				_val = nil
				g.storage.Set(ctx, key, emptyVal, opts.Expires.Missing.Duration)
				return nil, errNil
			}

			_bytes, e = conv.Encode(_val)
			if e != nil {
				panic(e)
			}
			g.storage.Set(ctx, key, _bytes, opts.Expires.Default.Duration)
			return _val, nil
		},
	)

	if e == nil {
		cpy := g.cpys[loaderName]
		if cpy == nil {
			_copy(dist, ret)
		} else {
			cpy(dist, ret)
		}
	}
	return e
}

func (g *Group) Do(ctx context.Context, loader string, dist interface{}, args NamedArgs) error {
	var err error
	var retry int
	opts := &g.Opts
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = g.do(ctx, loader, dist, args)
			if err == ErrRetryAfter && opts.RetryLimit > 0 && retry < opts.RetryLimit {
				retry++
				time.Sleep(opts.RetrySleep + time.Millisecond*time.Duration(rand.Int()%15))
				continue
			}
			if err == errNil {
				return ErrNotFound
			}
			return err
		}
	}
}

// dist must be a pointer
func _copy(dist, src interface{}) {
	rv := reflect.ValueOf(src)
	if rv.Kind() == reflect.Ptr {
		if dist == src {
			return
		}
		reflect.ValueOf(dist).Elem().Set(rv.Elem())
		return
	}
	reflect.ValueOf(dist).Elem().Set(rv)
}
