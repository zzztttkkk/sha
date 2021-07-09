package sha

import (
	"fmt"
	"testing"
)

func Test_CookieJar(t *testing.T) {
	jar := NewCookieJar()
	_ = jar.Update("a.com", `q=qwww; domain=a.com; maxage=99999`)
	_ = jar.Update("a.com", `a=qwww; domain=api.a.com`)
	_ = jar.Update("a.com", `b=qwww; domain=api.b.com`)
	_ = jar.Update("a.com", `c=qwww; domain=api.c.com`)
	_ = jar.Update("a.com", "d=qwww; domain=v.a.com")
	_ = jar.Update("a.com", "d=qwww; domain=v.api.a.com")

	fmt.Println(jar.Cookies("v.api.a.com"))
	fmt.Println(jar.Cookies("a.com"))
	fmt.Println(jar.Cookies("audio.a.com"))

	_ = jar.SaveTo("./a.cookies.json")
}
