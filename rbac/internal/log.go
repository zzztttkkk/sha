package internal

import (
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	Logger = log.New(os.Stderr, "sha.rbac ", log.LstdFlags)
}
