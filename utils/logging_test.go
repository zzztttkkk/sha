package utils

import (
	"log"
	"testing"
)

func TestAcquireGroupLogger(t *testing.T) {
	l := AcquireGroupLogger("xx")
	l.Println("asdasd")
	l.Println("dfsdfsdf")

	log.SetOutput(&DailyOutput{prefix: "./"})

	l.Free()
}
