package pipeline

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"testing"
)

func TestNewRedisBackend(t *testing.T) {
	NewPipeline("a", nil, "A")

	backend := NewRedisBackend("sha:pipeline", redis.NewClient(&redis.Options{Addr: "127.0.0.1:16379"}))
	task, err := backend.PushTask(context.Background(), "A", 10, 12, 0)
	fmt.Println(task, err)
	fmt.Println(backend.PopTask(context.Background()))
}
