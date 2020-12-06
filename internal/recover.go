package internal

func Silence(fn func()) {
	defer func() { recover() }()
	fn()
}
