package captcha

import (
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/golang/freetype/truetype"
)

func init() {
	RegisterFont("微软雅黑", "C:/Windows/Fonts/simkai.ttf", &truetype.Options{Size: 32})
}

func TestNewImage(t *testing.T) {
	img := RenderString(
		"微软雅黑",
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
}

func BenchmarkRender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		img := RenderString(
			"微软雅黑",
			"我可以吞下玻璃而不伤身体",
			&Option{
				OffsetX: 10, OffsetY: 10,
				Color:          color.Black,
				Points:         200,
				AsciiHalfWidth: true,
			},
		)
		of, _ := os.OpenFile("a.png", os.O_WRONLY|os.O_CREATE, 0766)
		_ = png.Encode(of, img)
	}
}
