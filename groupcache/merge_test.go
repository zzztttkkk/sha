package groupcache

import (
	"fmt"
	"testing"
)

func TestMerge(_ *testing.T) {
	var opts Options
	group := New("a", &opts)
	fmt.Println(group.Opts)
}
