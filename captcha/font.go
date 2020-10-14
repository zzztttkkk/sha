package captcha

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

type _Face struct {
	font.Face
	size int
}

type _FontCache struct {
	sync.RWMutex

	all map[string]*_Face
}

var fontCache _FontCache

func init() {
	fontCache.all = map[string]*_Face{}
}

func RegisterFont(name, file string, options *truetype.Options) {
	fontCache.Lock()
	defer fontCache.Unlock()

	if _, exists := fontCache.all[name]; exists {
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

	face := &_Face{Face: truetype.NewFace(ttf, options)}
	if options == nil {
		face.size = 12
	} else {
		face.size = int(options.Size)
	}
	fontCache.all[name] = face
}

func getFace(name string) *_Face {
	fontCache.RLock()
	defer fontCache.RUnlock()
	return fontCache.all[name]
}
