package sha

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCli(t *testing.T) {
	cli := NewCli(nil)

	var wg = &sync.WaitGroup{}
	for i := 0; i < 50; i++ {
		go func() {
			wg.Add(1)
			defer wg.Done()

			baseCtx, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)
			defer cancelFunc()

			ctx := AcquireRequestCtx(baseCtx)
			defer ReleaseRequestCtx(ctx)

			req := &ctx.Request
			req.SetMethod(MethodGet)
			req.SetPathString("/")

			cli.Send(ctx, "www.baidu.com:443", true, func(_ *CliSession, err error) {
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("Req: %d %s\r\n", ctx.Response.StatusCode(), ctx.Response.Phrase())
			})
		}()
	}

	wg.Wait()
}
