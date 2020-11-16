package suna

import (
	"fmt"
	"testing"
)

func Test_i(t *testing.T) {
	var a = []byte("\r\n   ")
	fmt.Println(len(inplaceTrimSpace(a)))
}
