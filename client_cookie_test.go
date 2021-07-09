package sha

import (
	"context"
	"fmt"
	"testing"
)

func TestClientCookie(_ *testing.T) {
	opts := &CliOptions{
		CookiePersistence: "./cookies.json",
	}

	cli := NewCli(opts)
	defer cli.Close()

	rctx := AcquireRequestCtx(context.Background())
	defer ReleaseRequestCtx(rctx)

	req := &rctx.Request
	req.SetMethod(MethodGet)
	req.SetPath([]byte("/"))
	_ = cli.Send(rctx, "https://www.baidu.com")
	for _, v := range rctx.Response.Header().GetAll(HeaderSetCookie) {
		fmt.Println(string(v))
	}
	fmt.Println(cli.jar.all)
}
