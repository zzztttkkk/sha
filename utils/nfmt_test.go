package utils

import "testing"

func TestZFmt(t *testing.T) {
	zf := NewNamedFmt("1 { a: 04d }, 2{b}, 3{{c}, 4 {{}}, 5 {}}", true)

	println(zf.Render(M{"a": 45, "b": false}))
}
