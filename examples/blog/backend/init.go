package backend

import "github.com/zzztttkkk/snow/examples/blog/backend/internal"

func Init() {
	internal.LazyExecutor.Execute(nil)
}
