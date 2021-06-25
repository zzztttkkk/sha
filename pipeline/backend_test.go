package pipeline

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)

func TestNewRedisBackend(t *testing.T) {
	NewPipeline(
		"a", "A", HandlerFunc(func(ctx context.Context, task *Task, prevResult interface{}) (interface{}, error) {
			fmt.Println(task.id)
			time.Sleep(time.Second)
			return nil, nil
		}),
	)
	rcli := redis.NewClient(&redis.Options{Addr: "127.0.0.1:16379"})
	rcli.FlushDB(context.Background())

	Init(NewRedisBackend("sha:pipeline", rcli), context.Background(), 5)

	go func() {
		for {
			_, _ = Push(context.Background(), "A", 10, time.Now().Unix(), 0)
			time.Sleep(time.Millisecond * 800)
		}
	}()

	Start(context.Background())
}
