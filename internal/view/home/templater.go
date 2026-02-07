package home

import (
	"context"
	"io"
)

type Templater struct {}

func NewTemplater() *Templater{
	return &Templater{}
}

func (t *Templater) Home(ctx context.Context, w io.Writer) error {
	return home().Render(ctx, w)
}

