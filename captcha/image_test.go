package captcha

import (
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/golang/freetype/truetype"
)

func init() {
	RegisterFont("微软雅黑32", "C:/Windows/Fonts/simkai.ttf", &truetype.Options{Size: 32})
	RegisterFont("华文行楷32", "C:/Windows/Fonts/STXINGKA.TTF", &truetype.Options{Size: 32})
	RegisterFont("微软雅黑24", "C:/Windows/Fonts/simkai.ttf", &truetype.Options{Size: 24})
	RegisterFont("华文行楷24", "C:/Windows/Fonts/STXINGKA.TTF", &truetype.Options{Size: 24})
}

func TestNewImage(t *testing.T) {
	img := RenderOneFont(
		"*",
		"我可以吞下玻璃而不伤身体",
		&Option{
			OffsetX: 10, OffsetY: 10,
			Color:          color.Black,
			Points:         200,
			AsciiHalfWidth: true,
			Shuffle:        true,
		},
	)

	of, _ := os.OpenFile("a.png", os.O_WRONLY|os.O_CREATE, 0766)
	_ = png.Encode(of, img)

	img = RenderSomeFonts(
		-1,
		"我可以吞下玻璃而不伤身体",
		&Option{
			OffsetX: 10, OffsetY: 10,
			Color:          color.Black,
			Points:         200,
			AsciiHalfWidth: true,
			Shuffle:        true,
		},
	)

	of, _ = os.OpenFile("b.png", os.O_WRONLY|os.O_CREATE, 0766)
	_ = png.Encode(of, img)
}
