package sha

import (
	"fmt"
	"testing"
)

func Test_CookieJar(t *testing.T) {
	jar := NewCookieJar()
	err := jar.Update(
		"v2ex.com",
		`a="1222"; expires=Sun, 03 Oct 2021 14:41:24 GMT; httponly; Path=/; domain=aa.v2ex.com`,
	)
	fmt.Println(err)
	fmt.Println(jar.all)
}
