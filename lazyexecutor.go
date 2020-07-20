package suna

import (
	"fmt"
	"sort"
)

type Kwargs map[string]interface{}

type _Func struct {
	raw      func(args Kwargs)
	priority int
}

type _Funcs []*_Func

func (lst _Funcs) Len() int {
	return len(lst)
}

func (lst _Funcs) Less(i, j int) bool {
	return lst[i].priority < lst[j].priority
}

func (lst _Funcs) Swap(i, j int) {
	lst[i], lst[j] = lst[j], lst[i]
}

type _LazyExecutorT struct {
	lst _Funcs
}

func (executor *_LazyExecutorT) Register(fn func(kwargs Kwargs)) {
	executor.lst = append(executor.lst, &_Func{raw: fn, priority: 0})
}

func (executor *_LazyExecutorT) RegisterWithPriority(fn func(kwargs Kwargs), priority *Priority) {
	executor.lst = append(executor.lst, &_Func{raw: fn, priority: priority.base})
}

func (executor *_LazyExecutorT) Execute(kwargs Kwargs) {
	sort.Sort(executor.lst)

	for _, fn := range executor.lst {
		fn.raw(kwargs)
	}
}

func NewLazyExecutor() *_LazyExecutorT {
	return &_LazyExecutorT{}
}

type Priority struct {
	base int
	step int
}

func NewPriority(base int) *Priority {
	return &Priority{base: base}
}

func (p *Priority) Copy() *Priority {
	return &Priority{base: p.base, step: p.step}
}

func (p *Priority) Incr() *Priority {
	p.step++
	return &Priority{
		base: p.base + p.step,
	}
}

func (p *Priority) String() string {
	return fmt.Sprintf("Priority <%d>", p.base)
}
