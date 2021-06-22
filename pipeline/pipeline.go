package pipeline

import (
	"context"
	"errors"
	"sync"
)

type Pipeline struct {
	isHead  bool
	name    string
	handler Handler
	next    *Pipeline
}

type nothing struct {
	v int
}

var _begin interface{} = &nothing{0}

func IsBegin(v interface{}) bool { return v == _begin }

var ErrCancled = errors.New("sha.pipeline: task is canceled")

func (pl *Pipeline) Process(ctx context.Context, task *Task, prevResult interface{}) (interface{}, error) {
	if task.Status() == TaskStatusCanceld {
		return nil, ErrCancled
	}
	select {
	case <-ctx.Done():
		{
			return nil, ctx.Err()
		}
	default:
	}

	task.path = append(task.path, pl.name)
	v, err := pl.handler.Process(ctx, task, prevResult)
	if err != nil {
		return nil, err
	}
	if pl.next == nil {
		return v, nil
	}
	return pl.next.Process(ctx, task, v)
}

func (pl *Pipeline) AppendHandler(name string, handler Handler) *Pipeline {
	npl := &Pipeline{false, name, handler, nil}
	pl.next = npl
	return npl
}

var pipelineSync sync.Mutex
var pipelineMap = make(map[string]*Pipeline)

func NewPipeline(name string, handler Handler, taskType string) *Pipeline {
	pl := &Pipeline{false, name, handler, nil}

	pipelineSync.Lock()
	hpl := pipelineMap[taskType]
	if hpl == nil {
		pl.isHead = true
		pipelineMap[taskType] = pl
	} else {
		hpl.next = pl
	}
	pipelineSync.Unlock()
	return pl
}

func getPipeline(tType string) *Pipeline {
	pipelineSync.Lock()
	defer pipelineSync.Unlock()
	return pipelineMap[tType]
}
