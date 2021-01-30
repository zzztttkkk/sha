package utils

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestEncodeID(t *testing.T) {
	g := NewIDTokenGenerator(NewHashPoll(sha256.New, []byte("AAA")))
	fmt.Println(g.DecodeID(g.EncodeID(126)))
}
