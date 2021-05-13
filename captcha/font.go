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

type _CacheKey struct {
	v   rune
	dot fixed.Point26_6
}

type _CacheValue struct {
	dr      image.Rectangle
	mask    image.Image
	maskp   image.Point
	advance fixed.Int26_6
}

type _SyncCachedFace struct {
	sync.RWMutex
	font.Face

	size           int
	asciiHalfWidth bool
	cache          map[_CacheKey]*_CacheValue
}

func (f *_SyncCachedFace) Glyph(dot fixed.Point26_6, r rune) (image.Rectangle, image.Image, image.Point, fixed.Int26_6, bool) {
	f.RLock()
	key := _CacheKey{v: r, dot: dot}
	cached, ok := f.cache[key]
	if ok {
		if cached == nil {
			f.RUnlock()
			return image.Rectangle{}, nil, image.Point{}, 0, false
		}
		f.RUnlock()
		return cached.dr, cached.mask, cached.maskp, cached.advance, true
	}
	f.RUnlock()

	f.Lock()
	defer f.Unlock()

	dr, mask, maskp, advance, ok := f.Face.Glyph(dot, r)
	if ok {
		f.cache[key] = &_CacheValue{dr, mask, maskp, advance}
	} else {
		f.cache[key] = nil
	}
	return dr, mask, maskp, advance, ok
}

type _FontCache struct {
	m map[string]*_SyncCachedFace
	l []*_SyncCachedFace
}

var fontsCacheMutex sync.Mutex
var fontsCache = _FontCache{m: map[string]*_SyncCachedFace{}}

func RegisterFont(name, file string, options *truetype.Options, asciiHalfWidth bool) {
	fontsCacheMutex.Lock()
	defer fontsCacheMutex.Unlock()

	if _, exists := fontsCache.m[name]; exists {
		panic(fmt.Errorf("sha.captcha: font name(`%s`) file(`%s`) is already exists", name, file))
	}
	f, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	ttf, err := truetype.Parse(f)
	if err != nil {
		panic(err)
	}

	face := &_SyncCachedFace{
		Face:           truetype.NewFace(ttf, options),
		asciiHalfWidth: asciiHalfWidth,
		cache:          map[_CacheKey]*_CacheValue{},
	}
	if options == nil {
		face.size = 16
	} else {
		face.size = int(options.Size)
	}
	fontsCache.m[name] = face
	fontsCache.l = append(fontsCache.l, face)
}

func getFaceByName(name string) *_SyncCachedFace {
	if name == "*" {
		return fontsCache.l[int(rand.Uint32())%len(fontsCache.l)]
	}
	return fontsCache.m[name]
}

func getFaceByCount(i int) (l []*_SyncCachedFace) {
	if i < 1 || i >= len(fontsCache.l) {
		return fontsCache.l
	}
	for ; i > 0; i-- {
		l = append(l, fontsCache.l[int(rand.Uint32())%len(fontsCache.l)])
	}
	return
}
