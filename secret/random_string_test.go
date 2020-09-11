package secret

import (
	"fmt"
	"testing"
)

func TestRandBytes(_ *testing.T) {
	fmt.Println(string(RandBytes(10, nil)))
	fmt.Println(string(RandBytes(5, []byte("0123456789"))))
}
