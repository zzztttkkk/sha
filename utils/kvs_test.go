package utils

import (
	"fmt"
	"testing"
)

func TestAcquireutils.M(t *testing.T) {
	dict := AcquireKvs()
	dict.Append("a", 34)
	dict.Append("b", "5666")
	dict.Set("a", 56)
	dict.Remove("b")

	dict.EachNode(
		func(s string, i interface{}) {
			fmt.Println(s, i)
		},
	)
}
