package utils

import (
	"fmt"
	"testing"
)

type Y struct {
	O int
	P string
}

type X struct {
	Y Y
	A struct {
		X string
	}
	B struct {
		C int64
		D TomlDuration
	}
	E struct {
		F struct {
			G bool
		}
		K string
	}
}

func TestTomlFromFiles(t *testing.T) {
	var conf X
	TomlFromFiles(&conf, nil, "./toml_default.toml", "./toml_local.toml")
	fmt.Println(conf)
}
