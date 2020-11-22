package internal

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

func Stacks(v interface{}, skip int, size int) string {
	ptrs := make([]uintptr, size)
	i := runtime.Callers(skip, ptrs)
	frames := runtime.CallersFrames(ptrs[:i])

	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("Error: %v\n", v))

	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		buf.WriteString("\tFile: ")
		buf.WriteString(frame.File)
		buf.WriteString(" Line: ")
		buf.WriteString(strconv.FormatInt(int64(frame.Line), 10))
		buf.WriteString(" Func: ")
		buf.WriteString(frame.Function)
		buf.WriteRune('\n')
	}
	return buf.String()
}
