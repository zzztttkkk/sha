package utils

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestHashPool(t *testing.T) {
	poll := NewHashPoll(sha256.New, nil)

	for i := 0; i < 10; i++ {
		fmt.Println(string(poll.Sum([]byte("AAA"))))
	}
}
