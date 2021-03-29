package sha

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/sha/jsonx"
	"github.com/zzztttkkk/sha/utils"
	"testing"
	"time"
)

var headers = utils.MultiValueMap{
	"User-Agent":      []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36"},
	"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	"Accept-Language": []string{"zh-CN,zh;q=0.9,en;q=0.8"},
	"Cookie":          []string{"__bsi=17680847351351550412_00_73_N_N_0_0303_C02F_N_N_Y_0"},
}

func printResponse(ctx *RequestCtx) {
	ctx.Response.Header().EachItem(func(item *utils.KvItem) bool {
		fmt.Printf("%s: %s\n", item.Key, item.Val)
		return true
	})
	fmt.Printf("%s\n", ctx.Response.Body())
}

var rctxPool = NewRequestCtxPool(nil)

func TestConnection_Send(t *testing.T) {
	conn := rctxPool.NewHTTPSession("baidu.com:80", &Environment{Header: headers})
	var ctx RequestCtx
	ctx.Request.SetMethod("GET").SetPathString("/s")
	ctx.Request.Query().LoadMap(utils.MultiValueMap{"wd": []string{"qwer"}})

	err := conn.Send(&ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	printResponse(&ctx)
}

func TestConnection_TLS_Send(t *testing.T) {
	conn := rctxPool.NewHTTPSSession("baidu.com:443", &Environment{Header: headers})
	var ctx RequestCtx
	ctx.Request.SetMethod("GET").SetPathString("/s")
	ctx.Request.Query().LoadMap(utils.MultiValueMap{"wd": []string{"qwer"}})

	err := conn.Send(&ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	printResponse(&ctx)
}

func TestConnection_Send_Proxy(t *testing.T) {
	conn := rctxPool.NewHTTPSession("google.com:80", &Environment{Header: headers, HTTPProxy: HTTPProxy{Address: "127.0.0.1:56966"}})
	var ctx RequestCtx
	ctx.Request.SetMethod("GET").SetPathString("/s")
	ctx.Request.Query().LoadMap(utils.MultiValueMap{"wd": []string{"qwer"}})

	err := conn.Send(&ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	printResponse(&ctx)
}

func TestConnection_TLS_Send_Proxy(t *testing.T) {
	conn := rctxPool.NewHTTPSSession("google.com:443", &Environment{Header: headers, HTTPProxy: HTTPProxy{Address: "127.0.0.1:56966"}})
	var ctx RequestCtx
	ctx.Request.SetMethod("GET").SetPathString("/s")
	ctx.Request.Query().LoadMap(utils.MultiValueMap{"wd": []string{"qwer"}})

	err := conn.Send(&ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	printResponse(&ctx)
}

func TestGet(t *testing.T) {
	action := Get(context.Background(), "https://v1.hitokoto.cn/").KeepRequestCtx().Send()
	defer action.ReturnRequestCtx()

	if action.Err() != nil {
		fmt.Println(action.Err())
		return
	}
	res := action.Response()
	data, _ := jsonx.NewObject(res.Body().Bytes())

	fmt.Println(data.PeekStringDefault("", "hitokoto"))
}

func TestGetWithSession(t *testing.T) {
	session := DefaultRequestCtxPool().NewHTTPSSession("v1.hitokoto.cn:443", nil)
	_ = session.OpenConn(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	action := Get(ctx, "/").UseSession(session, "v1.hitokoto.cn").KeepRequestCtx().Send()
	defer action.Close()

	if action.Err() != nil {
		fmt.Println(action.Err())
		return
	}

	res := action.Response()
	data, _ := jsonx.NewObject(res.Body().Bytes())

	fmt.Println(data.PeekStringDefault("", "hitokoto"), action.CastTime())
}
