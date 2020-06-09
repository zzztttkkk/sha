package ini

import (
	"testing"
)

func TestLoad(t *testing.T) {
	Load("../examples/blog/conf.ini")
	Print()
}
