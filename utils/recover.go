package utils

func RecoverIf(fn func(interface{}) bool) {
	rv := recover()
	if rv == nil {
		return
	}
	if fn(rv) {
		return
	}
	panic(rv)
}
