package sha

import (
	"fmt"
	"testing"
)

func TestRequest(t *testing.T) {
	var req Request
	req.SetMethod(MethodConnect)
	req.SetPathString("www.google.com:443")
	req.parsePath()
	fmt.Println(&req.URL)

	req.URL.reset()
	req.SetMethod(MethodGet)
	req.SetPathString("/search?wd=45")
	req.parsePath()
	fmt.Println(&req.URL)
}
