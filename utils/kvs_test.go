package utils

import (
	"fmt"
	"testing"
)

func TestKvs(t *testing.T) {
	var kvs Kvs
	kvs.AppendString("a", "aa")
	kvs.Del("a")
	kvs.AppendString("b", "bb")
	fmt.Println(kvs.size, kvs.invalids, &kvs)
}
