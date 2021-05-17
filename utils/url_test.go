package utils

import (
	"fmt"
	"testing"
)

func TestEncodeURI(t *testing.T) {
	var buf []byte
	EncodeURI([]byte("Thyme &time=again"), &buf)
	fmt.Println(string(buf))
	buf = buf[:0]
	EncodeURIComponent([]byte("Thyme &time=again"), &buf)
	fmt.Println(string(buf))
}
