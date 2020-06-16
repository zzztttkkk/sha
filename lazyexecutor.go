package snow

type NamedArgs map[string]interface{}

type _LazyExecutorT struct {
	lst []func(args NamedArgs)
}

func (executor *_LazyExecutorT) Register(fn func(args NamedArgs)) {
	executor.lst = append(executor.lst, fn)
}

func (executor *_LazyExecutorT) Execute(args NamedArgs) {
	for _, fn := range executor.lst {
		fn(args)
	}
}

func NewLazyExecutor() *_LazyExecutorT {
	return &_LazyExecutorT{}
}
