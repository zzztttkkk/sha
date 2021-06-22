package pipeline

import (
	"sync/atomic"
	"time"
)

type TaskStatus int32

const (
	TaskStatusUnknown = TaskStatus(iota)
	TaskStatusInit
	TaskStatusRunning
	TaskStatusCanceld
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
	task.setStatus(TaskStatusCanceld)
	task.cleanUp()
}

func (task *Task) Err() error { return task.err }

func (task *Task) cleanUp() {
	runningLock.Lock()
	delete(runningMap, task.id)
	runningLock.Unlock()

	if task.cFn == nil {
		return
	}

	if atomic.LoadInt32(&task.cFnFlag) != 0 {
		return
	}
	atomic.StoreInt32(&task.cFnFlag, 1)
	task.cFn()
}
