package utils

import (
	"fmt"
	"testing"
)

func TestNewBitmap(t *testing.T) {
	bitmap := NewBitmap(128)
	bitmap.Add(127)
	bitmap.Add(120)
	bitmap.Add(111)

	fmt.Println(bitmap.String(0))
	fmt.Println(bitmap.String(1))

	a := NewBitmap(128)
	a.Add(1)
	a.Add(127)

	b := NewBitmap(10)
	b.Add(0)
	b.Add(9)
	b.Add(1)

	c := a.Or(b)
	d := a.And(b)

	fmt.Println(c.String(0))
	fmt.Println(c.String(1))
	fmt.Println(d.String(0))
}
