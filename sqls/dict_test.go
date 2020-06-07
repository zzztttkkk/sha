package sqls

import (
	"fmt"
	"testing"
)

func TestDict_ForUpdate(t *testing.T) {
	driverName = "postgres"

	dict := Dict{"s": 34, "j": 45}
	a, _ := dict.ForUpdate()
	fmt.Println(a)
}
