package pipeline

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"strings"
	"time"
)

type _RedisBackend struct {
	cli      redis.Cmdable
	prefix   string
	pushHash string
	popHash  string
}

type _RedisTask struct {
	Type    string        `json:"type"`
	Data    interface{}   `json:"data"`
	Timeout time.Duration `json:"timeout"`
}

var pushScript = `
local tid = redis.call("incr", "${prefix}:seq")
redis.call("zadd", "${prefix}:queue", ARGV[1], tid)
redis.call("hset", "${prefix}:task:"..tid, "data", ARGV[2], "id", tid)
return tid
`

var popScript = `
local v = {}

local tidLst = redis.call("zpopmin", "${prefix}:queue", 1)
if next(tidLst) == nil then
	return v
end

local tid = tidLst[1]
table.insert(v, tid)
table.insert(v, redis.call("hget", "${prefix}:task:"..tid, "data"))
return v
`

func (rb *_RedisBackend) PushTask(ctx context.Context, taskType string, priority int, data interface{}, timeout time.Duration) (id string, err error) {
	bs, err := jsonx.Marshal(_RedisTask{taskType, data, timeout})
	if err != nil {
		return "", err
	}
	v, e := rb.cli.EvalSha(ctx, rb.pushHash, nil, priority, bs).Result()
	if e != nil {
		return "", e
	}
	return strconv.FormatInt(v.(int64), 16), nil
}

func (rb *_RedisBackend) PopTask(ctx context.Context) (id string, taskType string, data interface{}, timeout time.Duration, err error) {
	v, err := rb.cli.EvalSha(ctx, rb.popHash, nil).Result()
	if err != nil {
		return "", "", nil, 0, err
	}

	lst := v.([]interface{})
	if len(lst) < 1 {
		return "", "", nil, 0, ErrEmpty
	}

	id = lst[0].(string)
	var task _RedisTask
	err = jsonx.Unmarshal(utils.B(lst[1].(string)), &task)
	if err != nil {
		return "", "", nil, 0, err
	}
	return id, task.Type, task.Data, task.Timeout, nil
}

func (rb *_RedisBackend) CancelTask(ctx context.Context, id string) error {
	panic("implement me")
}

func (rb *_RedisBackend) ReportResult(id string, duration time.Duration, err error, result interface{}) error {
	panic("implement me")
}

var _ Backend = (*_RedisBackend)(nil)

func NewRedisBackend(prefix string, cli redis.Cmdable) Backend {
	v := &_RedisBackend{cli: cli, prefix: prefix}

	h, e := cli.ScriptLoad(context.Background(), strings.ReplaceAll(pushScript, "${prefix}", prefix)).Result()
	if e != nil {
		panic(e)
	}
	v.pushHash = h

	h, e = cli.ScriptLoad(context.Background(), strings.ReplaceAll(popScript, "${prefix}", prefix)).Result()
	if e != nil {
		panic(e)
	}
	v.popHash = h

	return v
}
