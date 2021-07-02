package groupcache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"sync"
	"testing"
	"time"
	"unsafe"
)

type _AddArgs struct {
	A int
	B int
}

func (a *_AddArgs) Name() string {
	f := "%d+%d"
	if a.B < a.A {
		return fmt.Sprintf(f, a.B, a.A)
	}
	return fmt.Sprintf(f, a.A, a.B)
}

type Item struct {
	Sum int
}

var redisGroup = New("test", nil).SetRedisStorage(redis.NewClient(&redis.Options{DB: 7, Addr: "127.0.0.1:16379"})).RegisterLoader(
	"add-return-value",
	func(ctx context.Context, args NamedArgs) (ret interface{}, err error) {
		v := args.(*_AddArgs)
		time.Sleep(time.Second)
		fmt.Printf("!!!!!!!!!!!!!!Calc: %d %d\n", v.A, v.B)
		return Item{v.A + v.B}, err
	},
).RegisterLoader(
	"add-return-pointer",
	func(ctx context.Context, args NamedArgs) (ret interface{}, err error) {
		v := args.(*_AddArgs)
		time.Sleep(time.Second)
		fmt.Printf("--------------Calc: %d %d\n", v.A, v.B)
		return &Item{v.A + v.B}, err
	},
)

func init() {
	redisGroup.Opts.RetryLimit = 5
	redisGroup.Opts.RetrySleep.Duration = time.Millisecond * 300
}

func TestAddReturnValue(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(100)

	var arg = _AddArgs{A: rand.Int() % 100, B: rand.Int() % 100}

	for i := 0; i < 100; i++ {
		go func(ind int) {
			defer wg.Done()
			var ret *Item
			if err := redisGroup.Do(context.Background(), "add-return-pointer", &ret, &arg); err != nil {
				if err == ErrRetryAfter {
					fmt.Println("retry after")
					return
				}
			}
			fmt.Println(ind, ret, unsafe.Pointer(ret))
		}(i)
	}

	wg.Wait()
}
