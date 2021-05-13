package captcha

import (
	"fmt"
	"github.com/golang/freetype/truetype"
	"strconv"
	"strings"
)

// RegisterFonts
// FontName|FontFile
// FontName|FontFile|FontSize
// FontName|FontFile|FontSize|ASCII Half Width
func RegisterFonts(fonts ...string) {
	for _, f := range fonts {
		a := strings.Split(strings.TrimSpace(f), "|")
		for i, v := range a {
			a[i] = strings.TrimSpace(v)
		}

		var fn, ff string
		var fs float64 = 16
		var asciiHalfWidth bool

		switch len(a) {
		case 4:
			v, e := strconv.ParseFloat(a[2], 64)
			if e != nil {
				panic(e)
			}
			fn, ff, fs = a[0], a[1], v
			asciiHalfWidth = len(a[3]) > 0
		case 3:
			v, e := strconv.ParseFloat(a[2], 64)
			if e != nil {
				panic(e)
			}
			fn, ff, fs = a[0], a[1], v
		case 2:
			fn, ff = a[0], a[1]
		default:
			panic(fmt.Errorf("suna.captcha: error font config `%s`", f))
		}
		RegisterFont(fn, ff, &truetype.Options{Size: fs}, asciiHalfWidth)
	}
}
