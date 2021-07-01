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

var Marshal = jsonx.Marshal
var Unmarshal = jsonx.Unmarshal

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

type Options struct {
	Prefix     string             `toml:"prefix"`
	Expires    Expires            `toml:"expires"`
	RetryLimit int                `toml:"retry-limit"`
	RetrySleep utils.TomlDuration `toml:"retry-sleep"`
	MaxWait    int                `toml:"max-wait"`
}

type Group struct {
	Opts    Options
	name    string
	sf      *internal.SingleflightGroup
	loads   map[string]DataLoader
	cpys    map[string]func(dist, src interface{})
	storage Storage
}

var defaultGroupOpts = Options{
	Expires:    defaultExpires,
	RetrySleep: utils.TomlDuration{Duration: time.Millisecond * 50},
	MaxWait:    50,
}

func New(name string, opts *Options) *Group {
	g := &Group{loads: map[string]DataLoader{}, cpys: map[string]func(dist interface{}, src interface{}){}}
	g.name = name

	if opts == nil {
		g.Opts = defaultGroupOpts
	} else {
		g.Opts = *opts
		utils.Merge(&g.Opts, defaultGroupOpts)
	}

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

				err := Unmarshal(_bytes, dist)
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

			_bytes, e = Marshal(_val)
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
				time.Sleep(opts.RetrySleep.Duration + time.Millisecond*time.Duration(rand.Int()%15))
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
	dv := reflect.ValueOf(dist)
	dt := dv.Type()
	sv := reflect.ValueOf(src)
	st := sv.Type()

	if st.Kind() != reflect.Ptr { // int ==> *int
		dv.Elem().Set(sv)
		return
	}

	if st == dt { // *int ==> *int
		dv.Elem().Set(sv.Elem())
		return
	}

	if dt.Elem() == st { // *int ==> **int
		dv.Elem().Set(sv)
		return
	}

	if st.Elem() == dt { // **int ==> *int
		dv.Elem().Set(sv.Elem().Elem())
		return
	}
	panic(fmt.Errorf("sha.groupcache: can not copy `%v` to `%v`", src, dist))
}
