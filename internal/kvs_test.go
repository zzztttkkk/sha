package internal

import (
	"fmt"
	"testing"
)

func TestKvs(t *testing.T) {
	kvs := AcquireKvs()
	kvs.Append([]byte("A"), []byte("B"))
	ReleaseKvs(kvs)
	fmt.Println(kvs.invalidItems)

	kvs = AcquireKvs()
	fmt.Println(kvs.invalidItems)
	kvs.Append([]byte("A"), []byte("B"))
	fmt.Println(kvs.invalidItems)
	v, ok := kvs.Get([]byte("A"))
	fmt.Println(v, ok)
}
