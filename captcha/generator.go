package captcha

import (
	"context"
	"io"
)

type Generator interface {
	GenerateTo(ctx context.Context, w io.Writer) (string, error)
}
