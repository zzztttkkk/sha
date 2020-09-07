package utils

import (
	"fmt"
	"testing"
)

func TestFmt(t *testing.T) {
	f := NewNamedFmt("{a-x:06d} {b/a}")
	fmt.Println(f.Render(M{"a-x": 45, "b/a": 56}))
}
