package choretypes

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/models"
	"github.com/bpsoos/shiftbell/internal/view/layouts"
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

func (t *Templater) PageWithLayout(
	ctx context.Context,
	w io.Writer,
	offset int,
	limit int,
	chores *models.GetChoreTypeBatchResult,
) error {
	return layouts.Main().Render(templ.WithChildren(ctx, page(offset, limit, chores)), w)
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

func (t *Templater) Selector(
	ctx context.Context,
	w io.Writer,
	offset int,
	limit int,
	chores *models.GetChoreTypeBatchResult,
) error {
	return selector(offset, limit, chores).Render(ctx, w)
}
