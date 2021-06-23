package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Backend interface {
	PushTask(ctx context.Context, taskType string, priority int, data interface{}, timeout time.Duration) (id string, err error)
	PopTask(ctx context.Context) (id string, taskType string, data interface{}, timeout time.Duration, err error)
	CancelTask(ctx context.Context, id string) error
	ReportResult(id string, duration time.Duration, err error, result interface{}) error
}

var backendOnce sync.Once
var backend Backend

var runningLock sync.Mutex
var runningMap = make(map[string]*Task)

var DefaultWaitBufferSize = 12

var _baseCtx context.Context
var _maxWorkers int32

func Init(v Backend, baseCtx context.Context, maxWorkers int32) {
	backendOnce.Do(func() {
		backend = v
		if baseCtx == nil {
			baseCtx = context.Background()
		}
		_baseCtx = baseCtx
		_maxWorkers = maxWorkers
	})
}

func Push(ctx context.Context, taskType string, priority int, data interface{}, timeout time.Duration) (id string, err error) {
	if _, ok := pipelineMap[taskType]; !ok {
		return "", fmt.Errorf("sha.pipeline: unknown task type `%s`", taskType)
	}
	return backend.PushTask(ctx, taskType, priority, data, timeout)
}

var ErrEmpty = errors.New("sha.pipeline: empty")

func Pop(ctx context.Context) (task *Task, err error) {
	id, tType, data, timeout, err := backend.PopTask(ctx)
	if err != nil {
		return nil, err
	}

	task = &Task{id: id, data: data, status: TaskStatusInit, ctime: time.Now(), typeS: tType, timeout: timeout}

	runningLock.Lock()
	runningMap[task.id] = task
	runningLock.Unlock()

	return task, nil
}

func Cancel(ctx context.Context, id string) error {
	runningLock.Lock()
	task := runningMap[id]
	runningLock.Unlock()

	if task != nil {
		task.Cancel()
		return nil
	}
	return backend.CancelTask(ctx, id)
}

type UnexpectedError struct {
	inner interface{}
	path  []string
}

func (ue *UnexpectedError) Error() string {
	return fmt.Sprintf("sha.pipeline: unexpected error, `%v`", ue.inner)
}

func executeTask(task *Task) {
	pipeline := getPipeline(task.typeS)
	task.setStatus(TaskStatusRunning)

	ctx, cFn := context.WithCancel(_baseCtx)

	if task.timeout > 0 {
		var nCfn func()
		var oCfn = cFn
		ctx, nCfn = context.WithTimeout(ctx, task.timeout)
		cFn = func() {
			nCfn()
			oCfn()
		}
	}
	task.cFn = cFn

	var result interface{}
	var rErr error

	defer func() {
		rv := recover()
		if rv != nil {
			var ue = &UnexpectedError{rv, make([]string, len(task.path))}
			copy(ue.path, task.path)
			rErr = ue
		}

		task.setStatus(TaskStatusDone)
		_ = backend.ReportResult(task.id, time.Since(task.ctime), rErr, result)
	}()

	result, rErr = pipeline.Process(ctx, task, _begin)
}

var (
	MaxSleepDuration = time.Millisecond * 1000
	sleepStep        = time.Millisecond * 10
)

func Start(ctx context.Context) {
	var sleepDuration time.Duration
	var activeWorkers int32

	for {
		select {
		case <-ctx.Done():
			for {
				c := len(runningMap)
				if c > 0 {
					time.Sleep(time.Second)
					log.Printf("sha.pipeline: running tasks count: `%d`\r\n", c)
					continue
				}
				break
			}
			return
		default:
			if _maxWorkers > 0 && atomic.LoadInt32(&activeWorkers) > _maxWorkers {
				goto _doSleep
			}

			task, err := Pop(ctx)
			if err != nil {
				if err != ErrEmpty {
					log.Println(err)
				}
				goto _doSleep
			}

			if task.Status() != TaskStatusInit {
				task.Cancel()
				continue
			}
			sleepDuration = 0

			go func(_t *Task) {
				atomic.AddInt32(&activeWorkers, 1)
				defer atomic.AddInt32(&activeWorkers, -1)
				executeTask(_t)
			}(task)
		}

	_doSleep:
		sleepDuration += sleepStep
		if sleepDuration > MaxSleepDuration {
			sleepDuration = MaxSleepDuration
		}
		time.Sleep(sleepDuration)
		continue
	}
}

// PeekRunningTasks
// `fn` can not cancel the task, it will cause a deadlock.
// if you want do that, you should store the pointers of the tasks in another place and then cancel them after this function ends.
func PeekRunningTasks(fn func(*Task) bool) {
	runningLock.Lock()
	defer runningLock.Unlock()

	for _, v := range runningMap {
		if fn(v) {
			continue
		}
		break
	}
}
