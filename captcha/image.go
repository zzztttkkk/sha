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

type Option struct {
	Width          int
	Height         int
	OffsetX        int
	OffsetY        int
	Points         int
	Color          color.Color
	AsciiHalfWidth bool
	Shuffle        bool
}

var defaultFix = &Option{}

func newImageWithString(str []rune, faces []*_Face, option *Option) image.Image {
	fontsize := 0
	if len(faces) == 1 {
		fontsize = faces[0].size
	} else {
		for _, f := range faces {
			if f.size > fontsize {
				fontsize = f.size
			}
		}
	}

	var c color.Color
	if option == nil {
		option = defaultFix
		c = &color.Black
	} else {
		c = option.Color
	}

	dx := option.OffsetX
	if dx < 1 {
		dx = fontsize / 2
	}
	dy := option.OffsetY
	if dy < 1 {
		dy = fontsize / 2
	}

	w := option.Width
	if w < 1 {
		if option.AsciiHalfWidth {
			for _, r := range str {
				if r <= 255 { // ascii
					w += fontsize / 2
				} else {
					w += fontsize
				}
			}
		} else {
			w = len(str) * fontsize
		}
	}
	h := option.Height
	if h < 1 {
		h = fontsize
	}

	w += dx
	h += dy

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// points
	for n := option.Points; n > 0; n-- {
		a := rand.Int() % w
		b := rand.Int() % h

		img.Set(a, b, c)
		a++
		img.Set(a, b, c)
		b++
		img.Set(a, b, c)
		a--
		img.Set(a, b, c)
	}

	dot := fixed.P(0, fontsize)
	cursor := image.NewUniform(c)

	var dxV, dyV int
	for _, r := range str {
		dxV = dx
		dyV = dy
		if option.AsciiHalfWidth && r < 256 {
			dxV = dx / 2
			dyV = dy / 2
		}
		v := fixed.P(rand.Int()%dxV, rand.Int()%dyV)

		if rand.Uint32()&1 == 0 {
			v.X *= -1
		}
		if rand.Uint32()&1 == 0 {
			v.Y *= -1
		}

		dot.X += v.X
		dot.Y += v.Y

		var face *_Face
		if len(faces) == 1 {
			face = faces[0]
		} else {
			face = faces[rand.Int()%len(faces)]
		}

		dr, mask, maskp, advance, ok := face.Glyph(dot, r)
		if !ok {
			continue
		}
		draw.DrawMask(img, dr, cursor, image.Point{}, mask, maskp, draw.Over)
		dot.X += advance

		dot.X -= v.X
		dot.Y -= v.Y
	}
	return img
}

func RenderOneFont(fontname, txt string, option *Option) image.Image {
	var rs = []rune(txt)
	if option.Shuffle {
		rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })
	}
	return newImageWithString(rs, []*_Face{getFaceByName(fontname)}, option)
}

func RenderSomeFonts(count int, txt string, option *Option) image.Image {
	var rs = []rune(txt)
	if option.Shuffle {
		rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })
	}
	return newImageWithString(rs, getFaceByCount(count), option)
}
