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
	Points  int
	Color   color.Color
}

var defaultOption = &Options{}

func randFace(faces []*_Face) *_Face {
	l := len(faces)
	if l < 2 {
		return faces[0]
	}
	return faces[int(rand.Uint32())%l]
}

func newImageWithString(str []rune, faces []*_Face, option *Options) image.Image {
	if option == nil {
		option = defaultOption
	}

	type RuneAndFace struct {
		r  rune
		f  *_Face
		dx int
		dy int
	}

	c := option.Color
	if c == nil {
		c = color.RGBA{
			R: uint8(rand.Uint32() % 255),
			G: uint8(rand.Uint32() % 255),
			B: uint8(rand.Uint32() % 255),
			A: uint8(rand.Uint32()%55) + 200,
		}
	}

	var rfs []RuneAndFace
	w := 0
	fs := 0
	for _, v := range str {
		face := randFace(faces)
		dx := face.size / 3
		dy := dx

		if face.size > fs {
			fs = face.size
		}

		if face.asciiHalfWidth && v < 255 {
			w += face.size / 2
			dx = dx / 2
			dy = dy / 2
		} else {
			w += face.size
		}
		rfs = append(rfs, RuneAndFace{r: v, f: face, dx: dx, dy: dy})
	}
	h := fs + fs/2
	w += fs / 2
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// points
	for n := option.Points; n > 0; n-- {
		a := int(rand.Uint32()) % w
		b := int(rand.Uint32()) % h

		img.Set(a, b, c)
		a++
		img.Set(a, b, c)
		b++
		img.Set(a, b, c)
		a--
		img.Set(a, b, c)
	}

	dot := fixed.P(fs/4, fs)
	cursor := image.NewUniform(c)

	var dxV, dyV int
	var p image.Point
	for _, r := range rfs {
		dxV = r.dx
		dyV = r.dy
		v := fixed.P(int(rand.Uint32())%dxV, int(rand.Uint32())%dyV)
		if rand.Uint32()&1 == 0 {
			v.X *= -1
		}
		if rand.Uint32()&1 == 0 {
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

func RenderOneFont(fontname, txt string, option *Options) image.Image {
	return newImageWithString([]rune(txt), []*_Face{getFaceByName(fontname)}, option)
}

func RenderSomeFonts(count int, txt string, option *Options) image.Image {
	return newImageWithString([]rune(txt), getFaceByCount(count), option)
}
