package internal

import (
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	Logger = log.New(os.Stderr, "suna.rbac ", log.LstdFlags)
}
