package suna

import (
	"log"
	"os"
)

type Logger interface {
	Print(v ...interface{})
	Printf(f string, v ...interface{})
	Println(v ...interface{})
}

var logger Logger

func init() {
	logger = log.New(os.Stdout, "", log.LstdFlags)
}

func SetLogger(l Logger) {
	logger = l
}
