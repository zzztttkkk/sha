package utils

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

type groupLogger struct {
	name string
	buf  strings.Builder
}

func (l *groupLogger) output() {
	log.Print(l.buf.String())
}

func (l *groupLogger) Println(v ...interface{}) {
	l.buf.WriteByte('\t')
	_, _ = fmt.Fprintln(&l.buf, v...)
}

var groupLoggerPool = &sync.Pool{New: func() interface{} { return &groupLogger{} }}

func AcquireGroupLogger(name string) *groupLogger {
	v := groupLoggerPool.Get().(*groupLogger)
	v.name = name
	v.buf.WriteString(v.name)
	v.buf.WriteByte(':')
	v.buf.WriteByte('\n')
	return v
}

func ReleaseGroupLogger(logger *groupLogger) {
	logger.output()

	logger.buf.Reset()
	logger.name = ""
	groupLoggerPool.Put(logger)
}
