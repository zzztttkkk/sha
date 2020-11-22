package internal

import (
	"log"
	"os"
)

type Logger interface {
	Print(v ...interface{})
	Printf(f string, v ...interface{})
	Println(v ...interface{})
}

var L Logger

func init() {
	L = log.New(os.Stdout, "", log.LstdFlags)
}
