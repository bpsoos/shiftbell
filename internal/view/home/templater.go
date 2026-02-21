package home

import (
	"context"
	"io"

	"github.com/bpsoos/shiftbell/internal/models"
)

type Templater struct{}

func NewTemplater() *Templater {
	return &Templater{}
}

func (t *Templater) Home(offset int, limit int, ctx context.Context, w io.Writer) error {
	return home(offset, limit).Render(ctx, w)
}

func (t *Templater) GetChoreBatch(offset int, limit int, chores *models.GetChoreTypeBatchResult, ctx context.Context, w io.Writer) error {
	return getChoreBatch(offset, limit, chores).Render(ctx, w)
}
