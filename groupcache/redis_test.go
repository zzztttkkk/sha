package groupcache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type _AddArgs struct {
	A int
	B int
}

func (a *_AddArgs) Name() string {
	f := "%d:%d"
	if a.B < a.A {
		return fmt.Sprintf(f, a.B, a.A)
	}
	return fmt.Sprintf(f, a.A, a.B)
}

var c = New("test", nil, 10).SetRedisStorage(redis.NewClient(&redis.Options{DB: 7})).Append(
	"add-return-value",
	func(ctx context.Context, args NamedArgs) (ret interface{}, err error) {
		v := args.(*_AddArgs)
		time.Sleep(time.Second)
		fmt.Printf("Calc: %d %d\n", v.A, v.B)
		return v.A + v.B, err
	},
).Append(
	"add-return-pointer",
	func(ctx context.Context, args NamedArgs) (ret interface{}, err error) {
		v := args.(*_AddArgs)
		time.Sleep(time.Second)
		fmt.Printf("Calc: %d %d\n", v.A, v.B)
		r := v.A + v.B
		return &r, err
	},
)

func TestAddReturnValue(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(100)

	var arg = _AddArgs{A: rand.Int() % 100, B: rand.Int() % 100}

	for i := 0; i < 100; i++ {
		go func(ind int) {
			var ret int

			if err := c.Do(context.Background(), "add-return-value", &ret, &arg); err != nil {
				if err == ErrRetryAfter {
					fmt.Println("retry after")
					wg.Done()
					return
				}
			}
			fmt.Println(ind, ret)
			wg.Done()
		}(i)
	}

	wg.Wait()
}
