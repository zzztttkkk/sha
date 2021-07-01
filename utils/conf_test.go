package utils

import (
	"fmt"
	"testing"
)

func TestLoadTomlToMapInterface(t *testing.T) {
	var dist map[string]interface{}
	err := LoadToml(&dist, "./conf_test_1.toml", false)
	if err == nil {
		fmt.Println(dist)
		return
	}
	panic(err)
}

func TestLoadTomlToStruct(t *testing.T) {
	type Dist struct {
		StrSlice []string `toml:"str-slice"`
		VM       struct {
			GoPath string `toml:"gopath"`
			DDD    struct {
				XXX string `toml:"xxx"`
			} `toml:"ddd"`
		} `toml:"vm"`
	}

	var dist Dist
	err := LoadToml(&dist, "./conf_test_1.toml", false)
	if err == nil {
		fmt.Println(dist)
		return
	}
	panic(err)
}

func TestLoadTomlToStructPointer(t *testing.T) {
	type Dist struct {
		StrSlice []string `toml:"str-slice"`
		VM       *struct {
			GoPath string `toml:"gopath"`
			DDD    struct {
				XXX string `toml:"xxx"`
			} `toml:"ddd"`
		} `toml:"vm"`
	}

	var dist Dist
	err := LoadToml(&dist, "./conf_test_1.toml", false)
	if err == nil {
		fmt.Println(dist, *dist.VM)
		return
	}
	panic(err)
}

func TestRedis(t *testing.T) {
	var cfg = RedisConfig{Addrs: []string{"1270.0.0.1:16379"}}
	fmt.Println(cfg.Cli())
	fmt.Println(cfg.Mode)
}
