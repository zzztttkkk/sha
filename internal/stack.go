package internal

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

func LoadCallersFrames(v interface{}, skip int, size int) string {
	ptrs := make([]uintptr, size)
	i := runtime.Callers(skip, ptrs)
	frames := runtime.CallersFrames(ptrs[:i])

	buf := strings.Builder{}
	buf.WriteString(fmt.Sprintf("Error: %v; CallerFrames:\n", v))

	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		buf.WriteString("\t")
		buf.WriteString(frame.Function)
		buf.WriteString("\r\n\t\t")
		buf.WriteString(frame.File)
		buf.WriteByte(':')
		buf.WriteString(strconv.FormatInt(int64(frame.Line), 10))
		buf.WriteRune('\n')
	}
	return buf.String()
}
