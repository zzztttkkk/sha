package captcha

import (
	"fmt"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"io/ioutil"
	"math/rand"
	"sync"
)

type _Face struct {
	font.Face
	size           int
	asciiHalfWidth bool
	sync.Mutex
}

func (f *_Face) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	f.Lock()
	defer f.Unlock()

	return f.Face.Glyph(dot, r)
}

type _FontCache struct {
	m map[string]*_Face
	l []*_Face
}

var fontCache _FontCache

func init() {
	fontCache.m = map[string]*_Face{}
}

func RegisterFont(name, file string, options *truetype.Options, asciiHalfWidth bool) {
	if _, exists := fontCache.m[name]; exists {
		panic(fmt.Errorf("suna.captcha: font name(`%s`) file(`%s`) is already exists", name, file))
	}
	f, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	ttf, err := truetype.Parse(f)
	if err != nil {
		panic(err)
	}

	face := &_Face{Face: truetype.NewFace(ttf, options), asciiHalfWidth: asciiHalfWidth}
	if options == nil {
		face.size = 12
	} else {
		face.size = int(options.Size)
	}
	fontCache.m[name] = face
	fontCache.l = append(fontCache.l, face)
}

func getFaceByName(name string) *_Face {
	if name == "*" {
		return fontCache.l[int(rand.Uint32())%len(fontCache.l)]
	}
	return fontCache.m[name]
}

func getFaceByCount(i int) (l []*_Face) {
	if i < 0 || i >= len(fontCache.l) {
		return fontCache.l
	}
	for ; i > 0; i-- {
		l = append(l, fontCache.l[int(rand.Uint32())%len(fontCache.l)])
	}
	return
}
