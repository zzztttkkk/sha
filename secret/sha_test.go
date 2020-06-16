package secret

import (
	"fmt"
	"testing"
)

func TestSha(t *testing.T) {
	fmt.Println(string(Sha512.Calc([]byte(""))))
	fmt.Println(string(Md5.Calc([]byte(""))))
	fmt.Println(string(Default.Calc([]byte(""))))
}
