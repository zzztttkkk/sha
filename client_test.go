package sha

import (
	"context"
	"fmt"
	"testing"
)

func TestNewHTTPSession(t *testing.T) {
	session := NewHTTPSession("www.baidu.com", true, nil)
	_ = session.OpenConn(context.Background())
	defer session.Close()

	ctx := AcquireRequestCtx()
	defer ReleaseRequestCtx(ctx)

	ctx.Request.SetPathString("/s?wd=go")
	err := session.Send(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	res := &ctx.Response

	switch res.StatusCode() {
	case StatusOK:
		fmt.Println(res.Body().String())
	default:
		fmt.Println(res.StatusCode(), res.Phrase())
		fmt.Print(res.Header())
	}
}
