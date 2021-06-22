package pipeline

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Backend interface {
	Ping(context.Context) error
	PushTask(ctx context.Context, taskType string, data interface{}, timeout time.Duration) (id string, err error)
	PopTask(ctx context.Context) (id string, taskType string, data interface{}, timeout time.Duration, err error)
	CancelTask(ctx context.Context, id string) error
	ReportResult(id string, duration time.Duration, err error, result interface{}) error
}

var backendOnce sync.Once
var backend Backend

var runningLock sync.Mutex
var runningMap = make(map[string]*Task)

var DEFAULT_WAIT_BUFFER_SIZE = 12
var waitChan chan *Task

var _baseCtx context.Context

func Init(v Backend, waitBufferSize int, baseCtx context.Context) {
	backendOnce.Do(func() {
		backend = v
		if err := backend.Ping(context.Background()); err != nil {
			log.Fatalln(err)
		}

		if waitBufferSize < 0 {
			waitBufferSize = DEFAULT_WAIT_BUFFER_SIZE
		}
		waitChan = make(chan *Task, waitBufferSize)

		if baseCtx == nil {
			baseCtx = context.Background()
		}

		_baseCtx = baseCtx
	})
}

func Push(ctx context.Context, taskType string, data interface{}, timeout time.Duration) (id string, err error) {
	if _, ok := pipelineMap[taskType]; !ok {
		return "", fmt.Errorf("sha.pipeline: unknown task type `%s`", taskType)
	}
	return backend.PushTask(ctx, taskType, data, timeout)
}

func Pop(ctx context.Context) (task *Task, err error) {
	id, tType, data, timeout, err := backend.PopTask(ctx)
	if err != nil {
		return nil, err
	}

	task = &Task{id: id, data: data, status: TaskStatusInit, ctime: time.Now(), typeS: tType, timeout: timeout}

	runningLock.Lock()
	runningMap[task.id] = task
	runningLock.Unlock()

	waitChan <- task
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

type UnexpoctedError struct {
	inner interface{}
	path  []string
}

func (ue *UnexpoctedError) Error() string {
	return fmt.Sprintf("sha.pipeline: unexpected error, `%v`", ue.inner)
}

func Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				{
					return
				}
			default:
				{
					Pop(ctx)
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			{
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
			}
		case task := <-waitChan:
			{
				if task.Status() != TaskStatusInit {
					continue
				}

				go func() {
					pipeline := getPipeline(task.typeS)
					task.setStatus(TaskStatusRunning)

					ctx, cFn := context.WithCancel(_baseCtx)

					if task.timeout > 0 {
						var n_cFn func()
						var o_cFn = cFn
						ctx, n_cFn = context.WithTimeout(ctx, task.timeout)
						cFn = func() {
							n_cFn()
							o_cFn()
						}
					}
					task.cFn = cFn

					var result interface{}
					var rErr error

					defer func() {
						rv := recover()
						if rv != nil {
							var ue = &UnexpoctedError{rv, make([]string, len(task.path))}
							copy(ue.path, task.path)
							rErr = ue
						}

						task.setStatus(TaskStatusDone)
						backend.ReportResult(task.id, time.Since(task.ctime), rErr, result)

						task.cleanUp()
					}()

					result, rErr = pipeline.Process(ctx, task, _begin)
				}()
			}
		}
	}
}

// PeekRuningTasks
// `fn` can not cancel the task, it will cause a deadlock.
// if you want do that, you should store the pointers of the tasks in another place and then cancel them after this function ends.
func PeekRuningTasks(fn func(*Task) bool) {
	runningLock.Lock()
	defer runningLock.Unlock()

	for _, v := range runningMap {
		if fn(v) {
			continue
		}
		break
	}
}
