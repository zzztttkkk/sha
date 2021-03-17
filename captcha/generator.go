package captcha

import (
	"context"
	"image/png"
	"io"
	"math/rand"
)

type Generator interface {
	GenerateTo(ctx context.Context, w io.Writer) error
}

type TokenGenerator func(ctx context.Context) string
type TokensGenerator func(ctx context.Context) []string

type _RandTokenGenerator struct {
	tg       TokenGenerator
	fontname string
	opt      *Options
}

func (r *_RandTokenGenerator) GenerateTo(ctx context.Context, w io.Writer) error {
	return png.Encode(
		w,
		RenderOneFont(r.fontname, []rune(r.tg(ctx)), r.opt),
	)
}

func NewRandTokenGenerator(fn TokenGenerator, fontname string, opt *Options) Generator {
	if opt != nil {
		escapeV := *opt
		opt = &escapeV
	}
	return &_RandTokenGenerator{tg: fn, fontname: fontname, opt: opt}
}

type ShuffleOptions struct {
	Options
	PCT int
	Sep []rune
}

type _ShuffleStringGenerator struct {
	tg       TokensGenerator
	fontname string
	opt      *ShuffleOptions
}

func (r *_ShuffleStringGenerator) GenerateTo(ctx context.Context, w io.Writer) error {
	tokens := r.tg(ctx)
	var txt []rune
	var l = len(tokens)
	ind := rand.Int() % l
	for i, t := range tokens {
		if i == ind || rand.Int()%100 <= r.opt.PCT {
			p := []rune(t)
			rand.Shuffle(len(p), func(i, j int) { p[i], p[j] = p[j], p[i] })

			txt = append(txt, p...)
		} else {
			txt = append(txt, []rune(t)...)
		}

		if i < l {
			txt = append(txt, r.opt.Sep...)
		}
	}

	return png.Encode(
		w,
		RenderOneFont(r.fontname, txt, &r.opt.Options),
	)
}

func NewShuffleStringGenerator(fn TokensGenerator, fontname string, opt *ShuffleOptions) Generator {
	if opt != nil {
		escapeV := *opt
		opt = &escapeV
	}
	return &_ShuffleStringGenerator{tg: fn, fontname: fontname, opt: opt}
}
