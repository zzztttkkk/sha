package captcha

import "image"

type Provider interface {
	Provide() string
	GetOption() *Option
	GetFontName() string
}

func Render(provider Provider) (image.Image, string) {
	str := provider.Provide()
	return RenderString(provider.GetFontName(), str, provider.GetOption()), str
}
