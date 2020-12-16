# sha

the pinyin for `沙(sand)` is `shā`

# contents

- [server and router](#server)
	- [listen_and_serve](#listen_and_serve)
	- [router](#router)

- [validator](#validator)
	- [tags](#tags)
	- [validator_example](#validator_example)

- [sqlx](#sqlx)

# server

## listen_and_serve

```go
server := sha.Default(
	sha.RequestHandlerFunc(func(ctx *sha.RequestCtx) {
		_, _ = ctx.WriteString("Hello world!")
	}),
)
server.ListenAndServe()
```

## router

```go
mux := sha.NewMux("", nil)
mux.HTTP(
	"get", "/",
	sha.RequestHandlerFunc(func(ctx *sha.RequestCtx) {
		_, _ = ctx.WriteString("Hello world!")
	}),
)

// :name    match one part
// name:*   match any parts
mux.HTTP(
	"get", "/a/:id/points:*",
	sha.RequestHandlerFunc(func(ctx *sha.RequestCtx) {
		fmt.Println(ctx.Request.Params.Get("id"), ctx.Request.Params.Get("points"))
		_, _ = ctx.WriteString("Hello world!")
	}),
)

mux.WebSocket(
	"/ws",
	func(ctx context.Context, req *sha.Request, conn *websocket.Conn, _ string) {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			fmt.Println("data recoved: ", data)
		}
	},
)

mux.FileSystem(http.Dir("./"), "get", "/static/filename:*", true)

branchA := sha.NewBranch()
branchA.HTTP(
	"get", "/",
	sha.RequestHandlerFunc(func(ctx *sha.RequestCtx) {
		_, _ = ctx.WriteString("Hello world!")
	}),
)

mux.AddBranch("/a", branchA)

server := sha.Default(mux)
server.ListenAndServe()
```

# validator

validate the request form data by structure.

## tags

grammar: `[name]([,attrName[=attrVal]]*)`

```go
validator.RegisterRegexp("joinedints", regexp.MustCompile(`\d+(,\d+)*`))

type Form struct{
	Id int64 `validator:"id,V=1-20"` // int and in range[1,20]
	Ints string `validator:",R=joinedints"` // match regexp `joinedints`
	Strings []string `validator:"strings,S=3"` // form data is a string list and the list size is 3
}
```

tag attrs:

- P/p/params =>  peek data from path params, not form
- optional =>  set this field is optional
- disabletrimspace =>  disable trim space
- disableescapehtml =>  disable escape html in string/[]byte field
- description =>  set the description of this field, which show in document
- R/r/regexp =>  use custom regexp match data
- V/v/value =>	int value range[min-max]
- L/l/length => form data size range[min-max]
- S/s/size =>	list size range[min-max]

## validator_example

```go
// custom form field type
type Sha5256Hash []byte

func (pwd *Sha5256Hash) FormValue(v []byte) bool {
	n := sha512.New512_256()
	n.Write(v)
	dist := make([]byte, 64)
	hex.Encode(dist, n.Sum(nil))
	*pwd = dist
	return true
}

type TestForm struct {
	Name string  `validator:",L=3-20"`
	Nums []int64 `validator:",V=0-9,S=3"`
	Password Sha5256Hash
}

// return default value if the form data is empty
func (TestForm) Default(fieldName string) interface{} {
	switch fieldName {
	case "Nums":
		return []int64{1, 2, 3}
	}
	return nil
}

// use validator
handler := RequestHandlerFunc(func(ctx *RequestCtx) {
	var form Form
	ctx.MustValidate(&form)
	fmt.Printf("%+v\n", form)
	_, _ = ctx.WriteString("Hello World!")
})

// register the handler and the form together, then the api document will be automatically generated
mux.HTTPWithForm("post", "/form", handler, Form{})

// serving the document
mux.HandleDoc("get", "/doc")
```

# sqlx

this is a wrapper of [https://github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx).
