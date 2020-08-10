package secret

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSha(t *testing.T) {
	fmt.Println(string(Sha512.Calc([]byte("A"))))
	fmt.Println(string(Md5.Calc([]byte(""))))
	fmt.Println(string(_Default.Calc([]byte(""))))

	buf := bytes.NewBuffer([]byte("A"))
	fmt.Println(string(Sha512.StreamCalc(buf)))

	for i := 0; i < 100; i++ {
		fmt.Println(string(RandBytes(12, nil)))
	}
}
