package validator

var funcMap = map[string]func([]byte) ([]byte, bool){}
var funcDescp = map[string]string{}

func RegisterFunc(name string, fn func([]byte) ([]byte, bool), descp string) {
	funcMap[name] = fn
	funcDescp[name] = descp
}
