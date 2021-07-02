package sha

import (
	"context"
	"fmt"
	"testing"
)

func TestCliSession(t *testing.T) {
	cli := newCliSession(
		"www.google.com", true,
		&CliSessionOptions{
			HTTPProxy: HTTPProxy{Address: "127.0.0.1:51651"},
		},
	)

	ctx := AcquireRequestCtx(context.TODO())
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
	fmt.Printf("\r\nBodySize : %d\r\n", res.Body().Len())
	fmt.Println(res.Body().String())
}
