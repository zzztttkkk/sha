package pipeline

import "context"

type Handler interface {
	Process(context.Context, *Task, interface{}) (interface{}, error)
}

type HandlerFunc func(context.Context, *Task, interface{}) (interface{}, error)

func (fn HandlerFunc) Process(ctx context.Context, task *Task, prevResult interface{}) (interface{}, error) {
	return fn(ctx, task, prevResult)
}
