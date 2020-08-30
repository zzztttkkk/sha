package suna

import (
	"github.com/zzztttkkk/suna/config"
	"testing"
)

func TestInit(t *testing.T) {
	conf := config.Default()
	Init(&InitOption{Config: &conf})
}
