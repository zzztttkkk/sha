package session

import (
	"github.com/zzztttkkk/sha/utils"
	"testing"
)

func init() {
	Init(&Options{Redis: utils.RedisConfig{Addrs: []string{"127.0.0.1:16379"}}})
}

func TestNew(t *testing.T) {

}
