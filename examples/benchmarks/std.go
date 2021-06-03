package main

import "net/http"

type Std struct{}

func (_ Std) Name() string {
	return "std"
}

func (_ Std) HelloWorld(address string) {
	var message = []byte("HelloWorld!")
	_ = http.ListenAndServe(address, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(message)
	}))
}

var _ Engine = Std{}

func init() { register(Std{}) }
