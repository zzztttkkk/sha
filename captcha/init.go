package captcha

import (
	"fmt"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
	"github.com/zzztttkkk/suna/utils"
)

func init() {
	internal.Dig.LazyInvoke(
		func(conf *config.Suna) {
			for _, f := range conf.Captcha.Fonts {
				a := strings.Split(strings.TrimSpace(f), ":")
				var fn, ff string
				var fs float64 = 16
				switch len(a) {
				case 3:
					fn, ff, fs = a[0], a[1], float64(utils.S2I32(a[2]))
				case 2:
					fn, ff = a[0], a[1]
				default:
					panic(fmt.Errorf("suna.captcha: error font config `%s`", f))
				}
				RegisterFont(fn, ff, &truetype.Options{Size: fs})
			}
		},
	)
}
