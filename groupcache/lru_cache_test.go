package groupcache

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
	"unsafe"
)

var lruGroup = New("test", nil).SetStorage(NewLRUCache(100)).RegisterLoader(
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
	lruGroup.Opts.RetryLimit = 5
	lruGroup.Opts.RetrySleep.Duration = time.Millisecond * 300
}

func TestLru(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(100)

	var arg = _AddArgs{A: rand.Int() % 100, B: rand.Int() % 100}

	for i := 0; i < 100; i++ {
		go func(ind int) {
			defer wg.Done()
			var ret *Item
			if err := lruGroup.Do(context.Background(), "add-return-pointer", &ret, &arg); err != nil {
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
