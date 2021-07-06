package captcha

import (
	"fmt"
	"image/png"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/golang/freetype/truetype"
)

func init() {
	RegisterFont("微软雅黑", "C:/Windows/Fonts/simkai.ttf", &truetype.Options{Size: 64}, true)
}

func TestNewImage(t *testing.T) {
	img := RenderOneFont(
		"*",
		[]rune("21598"),
		&ImageOptions{
			OffsetX: 16, OffsetY: 10,
		},
	)

	of, _ := os.OpenFile("a.png", os.O_WRONLY|os.O_CREATE, 0766)
	_ = png.Encode(of, img)

	img = RenderSomeFonts(
		-1,
		[]rune("我可以吞下玻璃而不伤身体"),
		&ImageOptions{
			OffsetX: 16, OffsetY: 16,
		},
	)

	of, _ = os.OpenFile("b.png", os.O_WRONLY|os.O_CREATE, 0766)
	_ = png.Encode(of, img)
}

func removeAllPng() {
	files, _ := ioutil.ReadDir("./")
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".png") {
			_ = os.Remove(f.Name())
		}
	}
}

func TestRemoveAllPng(t *testing.T) {
	removeAllPng()
}

func TestConcurrency(t *testing.T) {
	removeAllPng()

	wg := &sync.WaitGroup{}
	wg.Add(1000)

	for i := 0; i < 1000; i++ {
		go func(ind int) {
			defer wg.Done()

			img := RenderOneFont(
				"*",
				[]rune("21598"),
				&ImageOptions{
					OffsetX: 16, OffsetY: 16,
				},
			)
			of, _ := os.OpenFile(fmt.Sprintf("test_concurrency_%d.png", ind), os.O_WRONLY|os.O_CREATE, 0766)

			if err := png.Encode(of, img); err != nil {
				fmt.Println(err)
			}
		}(i)
	}

	wg.Wait()
}
