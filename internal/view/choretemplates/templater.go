package choretemplates

import (
	"context"
	"io"

	"github.com/a-h/templ"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
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
	choreTemplates *models.GetChoreTemplateBatchResult,
) error {
	return page(offset, limit, choreTemplates).Render(ctx, w)
}

func (t *Templater) PageWithLayout(
	ctx context.Context,
	w io.Writer,
	offset int,
	limit int,
	choreTemplates *models.GetChoreTemplateBatchResult,
) error {
	return layouts.Main().Render(templ.WithChildren(ctx, page(offset, limit, choreTemplates)), w)
}

func (t *Templater) Table(
	ctx context.Context,
	w io.Writer,
	offset int,
	limit int,
	choreTemplates *models.GetChoreTemplateBatchResult,
) error {
	return table(offset, limit, choreTemplates).Render(ctx, w)
}

func (t *Templater) Selector(
	ctx context.Context,
	w io.Writer,
	offset int,
	limit int,
	choreTemplates *models.GetChoreTemplateBatchResult,
) error {
	return selector(offset, limit, choreTemplates).Render(ctx, w)
}
