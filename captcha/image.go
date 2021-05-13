package captcha

import (
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"time"

	"golang.org/x/image/math/fixed"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Options struct {
	OffsetX int
	OffsetY int
	Lines   int
	Color   color.Color
}

var defaultOption = &Options{}

func randFace(faces []*_SyncCachedFace) *_SyncCachedFace {
	l := len(faces)
	if l < 2 {
		return faces[0]
	}
	return faces[int(rand.Uint32())%l]
}

func newImageWithString(str []rune, faces []*_SyncCachedFace, option *Options) image.Image {
	if option == nil {
		option = defaultOption
	}

	type RuneAndFace struct {
		r  rune
		f  *_SyncCachedFace
		dx int
		dy int
	}

	// color has not effect on the machine
	c := option.Color
	if c == nil {
		c = color.RGBA{
			R: uint8(rand.Uint32() % 255),
			G: uint8(rand.Uint32() % 255),
			B: uint8(rand.Uint32() % 255),
			A: uint8(rand.Uint32()%55) + 200,
		}
	}

	var dx = int(float32(option.OffsetX+1)*1.5) / 2

	var rfs []RuneAndFace
	w := 0
	fs := 0
	for _, v := range str {
		face := randFace(faces)
		_dx := dx

		if face.size > fs {
			fs = face.size
		}

		if face.asciiHalfWidth && v < 255 {
			w += face.size / 2
			_dx = dx / 2
		} else {
			w += face.size
		}
		rfs = append(rfs, RuneAndFace{r: v, f: face, dx: _dx, dy: option.OffsetY})
	}
	h := fs + fs/2
	w += fs / 2
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	dot := fixed.P(fs/6, fs)
	cursor := image.NewUniform(c)

	var p image.Point
	for _, r := range rfs {
		v := fixed.P(rand.Int()%(r.dx)+r.dx, rand.Int()%(r.dy+1))
		if rand.Int()&1 == 0 {
			v.X *= -1
		}
		if rand.Int()&1 == 0 {
			v.Y *= -1
		}

		dot.X += v.X
		dot.Y += v.Y

		dr, mask, maskp, advance, ok := r.f.Glyph(dot, r.r)
		if !ok {
			continue
		}

		draw.DrawMask(img, dr, cursor, p, mask, maskp, draw.Over)

		dot.X += advance

		dot.X -= v.X
		dot.Y -= v.Y
	}

	return img
}

func RenderOneFont(fontname string, txt []rune, option *Options) image.Image {
	return newImageWithString(txt, []*_SyncCachedFace{getFaceByName(fontname)}, option)
}

func RenderSomeFonts(count int, txt []rune, option *Options) image.Image {
	return newImageWithString(txt, getFaceByCount(count), option)
}
