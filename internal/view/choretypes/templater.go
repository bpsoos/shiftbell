package choretypes

import (
	"context"
	"io"

	"github.com/bpsoos/shiftbell/internal/models"
)

type Templater struct{}

func NewTemplater() *Templater {
	return &Templater{}
}

func (t *Templater) Page(
	ctx context.Context,
	w io.Writer,
	offset int,
	limit int,
	chores *models.GetChoreTypeBatchResult,
) error {
	return page(offset, limit, chores).Render(ctx, w)
}

func (t *Templater) Table(
	ctx context.Context,
	w io.Writer,
	offset int,
	limit int,
	chores *models.GetChoreTypeBatchResult,
) error {
	return table(offset, limit, chores).Render(ctx, w)
}
