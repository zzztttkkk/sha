package secret

import (
	"fmt"
	"testing"
)

func TestRandBytes(_ *testing.T) {
	for i := 0; i < 100; i++ {
		fmt.Println(string(RandBytes(10, nil)))
		fmt.Println(string(RandBytes(5, []byte("0123456789"))))
	}
}
