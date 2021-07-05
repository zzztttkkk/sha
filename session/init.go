package session

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/zzztttkkk/sha/auth"
	"github.com/zzztttkkk/sha/captcha"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"hash/crc64"
	"math/bits"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type Options struct {
	Redis  utils.RedisConfig  `toml:"redis" json:"redis"`
	Prefix string             `json:"suffix" toml:"suffix"`
	MaxAge utils.TomlDuration `json:"max_age" toml:"max-age"`

	Captcha struct {
		MaxAge utils.TomlDuration `json:"max_age" toml:"max-age"`
		Skip   bool               `json:"skip" toml:"skip"`
	} `json:"captcha" toml:"captcha"`

	CSRF struct {
		CookieName string             `json:"cookie_name" toml:"cookie-name"`
		HeaderName string             `json:"header_name" toml:"header-name"`
		MaxAge     utils.TomlDuration `json:"max_age" toml:"max-age"`
		Skip       bool               `json:"skip" toml:"skip"`
	} `json:"csrf" toml:"csrf"`
}

var opts Options

func init() {
	utils.MathRandSeed()
}

var IDGenerator = func(v *[]byte) {
	const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	writeUint := func(q uint64) {
		var r uint64
		for {
			q, r = bits.Div64(0, q, 64)
			*v = append(*v, encodeStd[r])
			if q == 0 {
				break
			}
		}
	}
	writeUint(uint64(time.Now().UnixNano()))
	*v = append(*v, '.')
	for i := 0; i < 6; i++ {
		*v = append(*v, encodeStd[rand.Int()%64])
	}
}

var CRSFTokenGenerator = func(v []byte) {
	const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	for i := 0; i < len(v); i++ {
		v[i] = encodeStd[rand.Int()%64]
	}
}

var ImageCaptchaGenerator captcha.Generator
var AudioCaptchaGenerator captcha.Generator
var Marshal = jsonx.Marshal
var Unmarshal = jsonx.Unmarshal

var rcli redis.Cmdable
var once sync.Once
var DefaultOpts Options
var PrefixLength int
var maxage int64

func init() {
	DefaultOpts.Prefix = "session:"
	DefaultOpts.Redis.Mode = "singleton"
	DefaultOpts.MaxAge.Duration = time.Hour * 3
	DefaultOpts.Captcha.MaxAge.Duration = time.Minute * 5
	DefaultOpts.CSRF.CookieName = "csrf"
	DefaultOpts.CSRF.HeaderName = "X-CSRF"
	DefaultOpts.CSRF.MaxAge.Duration = time.Minute * 30
}

var (
	updateScriptHash       string
	clearScriptHash        string
	createScriptHash       string
	createWithUAScriptHash string
	deleteAllHash          string
)

func Init(opt *Options) {
	once.Do(func() {
		if opt == nil {
			opts = DefaultOpts
		} else {
			opts = *opt
			utils.Merge(&opts, DefaultOpts)
		}
		rcli = opts.Redis.Cli()
		PrefixLength = len(opt.Prefix)
		maxage = int64(opts.MaxAge.Duration / time.Second)

		var err error
		updateScriptHash, err = rcli.ScriptLoad(
			context.Background(),
			`
redis.call('hset', KEYS[1], KEYS[2], ARGV[1]);
redis.call('expire', KEYS[1], ARGV[2]);
return 1;
`,
		).Result()
		if err != nil {
			panic(err)
		}
		clearScriptHash, err = rcli.ScriptLoad(
			context.Background(),
			`
redis.call('del', KEYS[1]);
redis.call('hset', KEYS[1], '.created', redis.call('time')[1]);
return 1;
`,
		).Result()
		if err != nil {
			panic(err)
		}
		createScriptHash, err = rcli.ScriptLoad(
			context.Background(),
			`
redis.call('hset', KEYS[1], '.created', redis.call('time')[1]);
redis.call('expire', KEYS[1], ARGV[1]);
return 1;
`,
		).Result()
		if err != nil {
			panic(err)
		}
		createWithUAScriptHash, err = rcli.ScriptLoad(
			context.Background(),
			`
redis.call('hset', KEYS[1], '.created', redis.call('time')[1]);
redis.call('hset', KEYS[2], KEYS[3], KEYS[1])
redis.call('expire', KEYS[1], ARGV[1]);
return 1;
`,
		).Result()
		if err != nil {
			panic(err)
		}
		deleteAllHash, err = rcli.ScriptLoad(
			context.Background(),
			`
local keys = redis.call('hvals', KEYS[1]);
if #keys < 1 then
	return 0;
end
for idx=1, #keys do
	redis.call('del', keys[idx]);
end
redis.call('del', KEYS[1]);
return 1;
`,
		).Result()
		if err != nil {
			panic(err)
		}
	})
}

var ErrEmptySession = errors.New("sha.session: nil")

/*New make a session for `request`
1, do auth for `request`
2, if auth fails, get the session from `request` and check if it exists.
	2.1, if not exists, generate a new session and return
	2.2, if exists, return
3, try to get the prev session through user-agent
	3.1, if not exists, generate a new session and save it through user-agent, and then return
	3.2, if exists, return
*/
func New(ctx context.Context, request Request) (Session, error) {
	var err error
	ptr := request.GetSessionID()
	if ptr == nil {
		return nil, ErrEmptySession
	}

	subject, _ := auth.Auth(ctx)
	if subject == nil {
		if len(*ptr) > 0 {
			var temp = make([]byte, 0, 25)
			temp = append(temp, opts.Prefix...)
			temp = append(temp, *ptr...)
			*ptr = (*ptr)[:0]
			if rcli.TTL(ctx, utils.S(temp)).Val() > 1 {
				*ptr = append(*ptr, temp...)
				return *ptr, nil
			}
		}

		*ptr = append(*ptr, opts.Prefix...)
		IDGenerator(ptr)
		err = rcli.EvalSha(ctx, createScriptHash, []string{utils.S(*ptr)}, maxage).Err()
		if err != nil {
			return nil, err
		}
		request.SetSessionID()
		return *ptr, nil
	}

	var table crc64.Table
	hash := crc64.New(&table)
	_, _ = hash.Write(utils.B(request.UserAgent()))
	ua := strconv.FormatUint(hash.Sum64(), 16)
	rcli.HSet(ctx, fmt.Sprintf("%suseragents", opts.Prefix), ua, request.UserAgent())
	key := fmt.Sprintf("%suser:%s", opts.Prefix, subject.GetID())
	var sid string
	if sid, err = rcli.HGet(ctx, key, ua).Result(); err != nil && err != redis.Nil {
		return nil, err
	}
	if len(sid) > 0 {
		*ptr = append(*ptr, sid...)
		return *ptr, nil
	}
	*ptr = append(*ptr, opts.Prefix...)
	IDGenerator(ptr)
	err = rcli.EvalSha(ctx, createWithUAScriptHash, []string{utils.S(*ptr), key, ua}, maxage).Err()
	if err != nil {
		return nil, err
	}
	request.SetSessionID()
	return *ptr, nil
}

func Invalidate(ctx context.Context, subject auth.Subject) error {
	return rcli.EvalSha(ctx, deleteAllHash, []string{fmt.Sprintf("%suser:%s", opts.Prefix, subject.GetID())}).Err()
}
