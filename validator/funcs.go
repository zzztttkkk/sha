package validator

type ValidateFunc func([]byte) ([]byte, bool)

var funcMap = map[string]ValidateFunc{}
var funcDescp = map[string]string{}

// RegisterFunc you must call this function before calling other functions
func RegisterFunc(name string, fn ValidateFunc, descp string) {
	funcMap[name] = fn
	funcDescp[name] = descp
}
