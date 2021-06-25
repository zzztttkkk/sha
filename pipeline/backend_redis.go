package pipeline

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"strings"
	"time"
)

type _RedisBackend struct {
	cli        redis.Cmdable
	prefix     string
	pushHash   string
	popHash    string
	cancelHash string
	reportHash string
}

type _RedisTask struct {
	Type    string        `json:"type"`
	Data    interface{}   `json:"data"`
	Timeout time.Duration `json:"timeout"`
}

var pushScript = `
local tid = redis.call("incr", "${prefix}:seq")
redis.call("zadd", "${prefix}:queue", ARGV[1], tid)
local ctime = redis.call("TIME")
redis.call("hset", "${prefix}:task:"..tid, "data", ARGV[2], "id", tid, "ctime", ctime[1]..ctime[2])
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

var cancelScript = `
local key = "${prefix}:task:"..ARGV[1]
if redis.call("exists", key) == 0 then
	return 0
end

if redis.call("hexists", key, "etime") == 1 then
	return 1
end

redis.call("zrem", "${prefix}:queue", ARGV[1])
local ctime = redis.call("TIME")
redis.call("hset", key, "etime", ctime[1]..ctime[2], "error", "canceled")
return 2
`

var reportScript = `
local key = "${prefix}:task:"..ARGV[1]
if redis.call("exists", key) == 0 then
	return 0
end

if redis.call("hexists", key, "etime") == 1 then
	return 1
end

local ctime = redis.call("TIME")
redis.call("hset", key, "etime", ctime[1]..ctime[2], "error", ARGV[2], "result", ARGV[3], "duration", ARGV[4])
return 2
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

	lst, ok := v.([]interface{})
	if !ok || len(lst) < 1 {
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
	task := getRunningTask(id)
	if task != nil {
		task.Cancel()
		return nil
	}

	v, err := rb.cli.EvalSha(ctx, rb.cancelHash, nil, id).Result()
	if err != nil {
		return err
	}

	switch v.(int64) {
	case 0:
		return redis.Nil
	case 1:
		return fmt.Errorf("sha.pipeline: double report for task `%s`", id)
	default:
		return nil
	}
}

func (rb *_RedisBackend) ReportResult(id string, duration time.Duration, err error, result interface{}) error {
	var eStr string
	if err != nil {
		if err == ErrCanceled {
			eStr = "canceled"
		} else {
			eStr = err.Error()
			if eStr == "" {
				eStr = "empty error message"
			}
		}
	}
	var bs []byte
	if err == nil && result != nil {
		bs, err = jsonx.Marshal(result)
		if err != nil {
			return err
		}
	}
	v, err := rb.cli.EvalSha(context.Background(), rb.reportHash, nil, id, eStr, bs, int64(duration)).Result()
	if err != nil {
		return err
	}
	switch v.(int64) {
	case 0:
		return redis.Nil
	case 1:
		return fmt.Errorf("sha.pipeline: double report for task `%s`", id)
	default:
		return nil
	}
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

	h, e = cli.ScriptLoad(context.Background(), strings.ReplaceAll(cancelScript, "${prefix}", prefix)).Result()
	if e != nil {
		panic(e)
	}
	v.cancelHash = h

	h, e = cli.ScriptLoad(context.Background(), strings.ReplaceAll(reportScript, "${prefix}", prefix)).Result()
	if e != nil {
		panic(e)
	}
	v.reportHash = h
	return v
}
