package sha

import (
	"fmt"
	"net/url"
	"testing"
)

func TestA(t *testing.T) {
	var buf []byte
	encodeURI([]byte(";,/?:@&=+$#-_.!~*'()ABC abc 123我"), &buf)
	fmt.Println(string(buf))
	buf = buf[:0]
	encodeURIComponent([]byte(";,/?:@&=+$#-_.!~*'()ABC abc 123我"), &buf)
	fmt.Println(string(buf))

	fmt.Println(url.Parse("/aaa我?sdsd=34"))
}
