package snow

type _LazyExecutorT struct {
	lst []func(...interface{})
}

func (executor *_LazyExecutorT) Register(fn func(...interface{})) {
	executor.lst = append(executor.lst, fn)
}

func (executor *_LazyExecutorT) Execute(args ...interface{}) {
	for _, fn := range executor.lst {
		fn(args...)
	}
}

func NewLazyExecutor() *_LazyExecutorT {
	return &_LazyExecutorT{}
}
