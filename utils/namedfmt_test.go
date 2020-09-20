package utils

import (
	"fmt"
	"testing"
)

func TestFmt(t *testing.T) {
	f := NewNamedFmt("1${a去-x:06d}2 3${b/a}4 5{${c}}6 7{${dd}8")
	fmt.Println(f.Render(M{"a去-x": 45, "b/a": 56, "c": 4, "dd": 566}))
}
