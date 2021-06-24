package pipeline

import (
	"log"
	"sync/atomic"
	"time"
)

type TaskStatus int32

const (
	TaskStatusUnknown = TaskStatus(iota)
	TaskStatusInit
	TaskStatusRunning
	TaskStatusCanceled
	TaskStatusDone
)

type Task struct {
	status  TaskStatus
	path    []string
	err     error
	ctime   time.Time
	timeout time.Duration

	id    string
	typeS string
	data  interface{}

	cFnFlag int32
	cFn     func()
}

func (task *Task) Status() TaskStatus {
	return TaskStatus(atomic.LoadInt32((*int32)(&task.status)))
}

func (task *Task) setStatus(status TaskStatus) {
	atomic.StoreInt32((*int32)(&task.status), int32(status))
}

func (task *Task) Cancel() {
	if task.Status() == TaskStatusDone {
		return
	}
	task.setStatus(TaskStatusCanceled)
	task.cleanUp(ErrCanceled, nil)
}

func (task *Task) Err() error { return task.err }

func (task *Task) cleanUp(err error, result interface{}) {
	if atomic.LoadInt32(&task.cFnFlag) != 0 {
		return
	}
	atomic.StoreInt32(&task.cFnFlag, 1)

	delRunningTask(task.id)
	if err = backend.ReportResult(task.id, time.Since(task.ctime), err, result); err != nil {
		log.Printf("sha.pipeline: report error `%s`\r\n", err.Error())
	}

	if task.cFn != nil {
		task.cFn()
	}
}
