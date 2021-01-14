package captcha

import (
	"image/png"
	"os"
	"testing"

	"github.com/golang/freetype/truetype"
)

func init() {
	RegisterFont("微软雅黑32", "C:/Windows/Fonts/simkai.ttf", &truetype.Options{Size: 32}, true)
	RegisterFont("微软雅黑24", "C:/Windows/Fonts/simkai.ttf", &truetype.Options{Size: 24}, true)
}

func TestNewImage(t *testing.T) {
	img := RenderOneFont(
		"*",
		"21598",
		&Options{
			OffsetX: 10, OffsetY: 10,
			Points:  200,
			Shuffle: false,
		},
	)

	of, _ := os.OpenFile("a.png", os.O_WRONLY|os.O_CREATE, 0766)
	_ = png.Encode(of, img)

	img = RenderSomeFonts(
		-1,
		"我可以吞下玻璃而不伤身体",
		&Options{
			OffsetX: 10, OffsetY: 10,
			Points:  200,
			Shuffle: true,
		},
	)

	of, _ = os.OpenFile("b.png", os.O_WRONLY|os.O_CREATE, 0766)
	_ = png.Encode(of, img)
}
