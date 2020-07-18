package utils

import (
	"fmt"
	"os"
	"sync"
	"time"
)

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
