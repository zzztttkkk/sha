package session

import (
	"context"
	"github.com/zzztttkkk/sha/captcha"
	"github.com/zzztttkkk/sha/utils"
	"image/png"
	"io"
	"log"
)

type _RandBase58Generator struct{}

func (_ _RandBase58Generator) GenerateTo(_ context.Context, w io.Writer) (string, error) {
	buf := make([]rune, 6)
	utils.RandBase58Runes(buf)

	img := captcha.RenderSomeFonts(-1, buf, &opts.Captcha.ImgOptions)
	if err := png.Encode(w, img); err != nil {
		return "", err
	}
	return string(buf), nil
}

func _initCaptcha() {
	if len(opts.Captcha.Fonts) < 1 {
		log.Printf("sha.session: empty captcha image fonts")
		return
	}
	if opts.Captcha.ImgOptions.OffsetX < 1 {
		opts.Captcha.ImgOptions.OffsetX = 6
	}
	if opts.Captcha.ImgOptions.OffsetY < 1 {
		opts.Captcha.ImgOptions.OffsetY = 6
	}
	ImageCaptchaGenerator = _RandBase58Generator{}
	captcha.RegisterFonts(opts.Captcha.Fonts...)
}
