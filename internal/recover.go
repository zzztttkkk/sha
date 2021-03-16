package internal

import (
	"log"
)

var showSilenceError bool

func ShowSilenceError(v bool) {
	showSilenceError = v
}

func Silence(fn func()) {
	defer func() {
		v := recover()
		if showSilenceError && v != nil {
			log.Printf("%v\n", v)
		}
	}()
	fn()
}
