package utils

import "testing"

func TestAcquireGroupLogger(t *testing.T) {
	l := AcquireGroupLogger("xx")
	l.Println("asdasd")
	l.Println("dfsdfsdf")

	ReleaseGroupLogger(l)
}
