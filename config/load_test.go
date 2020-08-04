package config

import (
	"fmt"
	"testing"
)

func TestFromFiles(t *testing.T) {
	conf := FromFiles("./default.toml", "./local.toml")
	fmt.Printf("%+v\n", conf)
}
