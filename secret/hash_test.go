package secret

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSha(t *testing.T) {
	fmt.Println(string(Sha512.Calc([]byte("A"))))
	fmt.Println(string(Md5.Calc([]byte(""))))
	fmt.Println(string(Default.Calc([]byte(""))))

	buf := bytes.NewBuffer([]byte("A"))
	fmt.Println(string(Sha512.StreamCalc(buf)))
}
