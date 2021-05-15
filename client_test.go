package sha

import (
	"context"
	"fmt"
	"testing"
)

func TestNewHTTPSession(t *testing.T) {
	cli := NewHTTPClientSession(
		"www.google.com", true,
		&ClientSessionOptions{
			HTTPProxy: HTTPProxy{Address: "127.0.0.1:50266"},
		},
	)
	_ = cli.OpenConn(context.Background())
	defer cli.Close()

	ctx := AcquireRequestCtx()
	defer ReleaseRequestCtx(ctx)

	ctx.Request.SetPathString("/search?q=go")

	err := cli.Send(ctx)
	if err != nil {
		fmt.Println("Err: ", err)
		return
	}

	res := &ctx.Response

	fmt.Println(res.StatusCode(), res.Phrase())
	fmt.Print(res.Header())
	fmt.Println("")
	fmt.Println(res.Body().String())
}
