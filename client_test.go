package sha

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestCli(t *testing.T) {
	cli := NewCli(nil)
	defer cli.Close()

	var wg = &sync.WaitGroup{}
	for i := 0; i < 50; i++ {
		go func() {
			wg.Add(1)
			defer wg.Done()

			baseCtx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*250)
			defer cancelFunc()

			ctx := AcquireRequestCtx(baseCtx)
			defer ReleaseRequestCtx(ctx)

			req := &ctx.Request
			req.SetMethod(MethodGet)
			req.SetPathString("/")

			cli.Send(ctx, "https://www.baidu.com:443", func(_ *CliSession, err error) {
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("Res (%s): %d %s\r\n", ctx.TimeSpent(), ctx.Response.StatusCode(), ctx.Response.Phrase())
			})
		}()
	}

	wg.Wait()
}

func TestCliRedirect(t *testing.T) {
	go func() {
		ListenAndServe("127.0.0.1:5986", RequestHandlerFunc(func(ctx *RequestCtx) {
			num, _ := strconv.ParseInt(string(ctx.Request.Path()[1:]), 10, 32)
			if num < 100 {
				ctx.Response.SetStatusCode(StatusMovedPermanently)
				ctx.Response.Header().SetString(HeaderLocation, fmt.Sprintf("/%d?time=%d", num+1, time.Now().UnixNano()))
				return
			}
			ctx.Response.SetStatusCode(StatusOK)
			_ = ctx.WriteString("OK!")
		}))
	}()

	time.Sleep(time.Second)

	cli := NewCli(nil)
	cli.Opts.MaxRedirect = 100
	cli.Opts.KeepRedirectHistory = true
	defer cli.Close()

	ctx := AcquireRequestCtx(context.Background())
	defer ReleaseRequestCtx(ctx)

	ctx.Request.SetPathString("/0")
	cli.Send(ctx, "127.0.0.1:5986", func(_ *CliSession, err error) {
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Res: %d %s %s\r\n", ctx.Response.StatusCode(), ctx.Response.Phrase(), ctx.Request.History())
	})
}

func TestCliRedirectToAnotherHost(t *testing.T) {
	go func() {
		ListenAndServe("", RequestHandlerFunc(func(ctx *RequestCtx) {
			num, _ := strconv.ParseInt(string(ctx.Request.Path()[1:]), 10, 32)
			if num < 100 {
				ctx.Response.SetStatusCode(StatusMovedPermanently)
				ctx.Response.Header().SetString(
					HeaderLocation,
					fmt.Sprintf("/%d?time=%d", num+1, time.Now().UnixNano()),
				)
				return
			}
			ctx.Response.SetStatusCode(StatusMovedPermanently)
			ctx.Response.Header().SetString(
				HeaderLocation,
				fmt.Sprintf("https://www.baidu.com/aaaa?time=%d", time.Now().UnixNano()),
			)
			return
		}))
	}()

	time.Sleep(time.Second)

	cli := NewCli(nil)
	cli.Opts.MaxRedirect = 101
	cli.Opts.KeepRedirectHistory = true
	defer cli.Close()

	ctx := AcquireRequestCtx(context.Background())
	defer ReleaseRequestCtx(ctx)

	ctx.Request.SetPathString("/0")
	cli.Send(ctx, "127.0.0.1:5986", func(_ *CliSession, err error) {
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf(
			"Res (%s): %d %s %s\r\n",
			ctx.TimeSpent(), ctx.Response.StatusCode(), ctx.Response.Phrase(), ctx.Request.History(),
		)
	})
}
