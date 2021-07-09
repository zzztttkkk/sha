package sha

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

const chromeUA = `
Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36
`

func TestClientCookie(_ *testing.T) {
	opts := &CliOptions{CookieStoragePath: "./cookies.json"}
	opts.BeforeSendRequest = append(
		opts.BeforeSendRequest,
		func(ctx *RequestCtx, host string) error {
			ctx.Request.Header().AppendString(HeaderUserAgent, strings.TrimSpace(chromeUA))
			return nil
		},
	)
	opts.AfterReceiveResponse = append(
		opts.AfterReceiveResponse,
		func(ctx *RequestCtx, err error) {
			for _, v := range ctx.Response.Header().GetAll(HeaderSetCookie) {
				fmt.Println(string(v))
			}
		},
	)

	cli := NewCli(opts)
	defer cli.Close()

	rctx := AcquireRequestCtx(context.Background())
	defer ReleaseRequestCtx(rctx)

	req := &rctx.Request
	req.SetMethod(MethodGet)
	req.SetPath([]byte("/"))
	_ = cli.Send(rctx, "https://www.baidu.com")
	fmt.Println(cli.jar.all)
}
