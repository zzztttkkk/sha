package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
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

type DailyOutput struct {
	file   *os.File
	mutex  sync.Mutex
	date   int
	prefix string
}

func (output *DailyOutput) ensure() {
	output.mutex.Lock()
	defer output.mutex.Unlock()

	y, m, d := time.Now().Date()
	if d == output.date {
		return
	}

	file, err := os.OpenFile(
		fmt.Sprintf("%s%d-%02d-%02d.log", output.prefix, y, m, d),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0777,
	)
	if err != nil {
		panic(err)
	}

	output.date = d
	output.file = file
}

func (output *DailyOutput) Write(v []byte) (int, error) {
	output.ensure()
	return output.file.Write(v)
}
