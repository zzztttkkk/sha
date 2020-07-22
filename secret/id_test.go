package secret

import (
	"fmt"
	"testing"
)

func TestDumpId(t *testing.T) {
	fmt.Println(DumpId(787878, 8640000000))
	fmt.Println(LoadId(DumpId(11111, 0)))
}
