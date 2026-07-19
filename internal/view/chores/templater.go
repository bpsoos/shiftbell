package chores

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/logging"
	"github.com/bpsoos/shiftbell/internal/models"
	"github.com/bpsoos/shiftbell/internal/view/layouts"
)

type Templater struct{}

func NewTemplater() *Templater {
	return &Templater{}
}

func (t *Templater) Table(
	ctx context.Context,
	w io.Writer,
	offset int,
	limit int,
	chores *models.GetChoreBatchResult,
) error {
	return table(offset, limit, chores).Render(ctx, w)
}

func (t *Templater) Page(
	ctx context.Context,
	w io.Writer,
	offset int,
	limit int,
	chores *models.GetChoreBatchResult,
	choreTypes *models.GetChoreTypeBatchResult,
	selectedChoreType *models.ChoreType,
) error {
	return page(offset, limit, chores, choreTypes, selectedChoreType).Render(ctx, w)
}

func (t *Templater) Chore(
	ctx context.Context,
	w io.Writer,
	chore *models.Chore,
) error {
	return choreCard(chore).Render(ctx, w)
}

func (t *Templater) NewChorePage(
	ctx context.Context,
	w io.Writer,
) error {
	return layouts.Main().Render(
		templ.WithChildren(
			ctx,
			newChorePage(),
		),
		w,
	)
}

func (t *Templater) PageWithLayout(
	ctx context.Context,
	w io.Writer,
	offset int,
	limit int,
	chores *models.GetChoreBatchResult,
) error {
	return layouts.Main().Render(
		templ.WithChildren(
			ctx,
			page(offset, limit, chores, nil, nil),
		),
		w,
	)
}

func (t *Templater) JoinedComponents(
	ctx context.Context,
	w io.Writer,
	componentSpecifiers ...models.NewChoreTypeComponent,
) error {
	components := make([]templ.Component, 0)
	for i := range componentSpecifiers {
		attrs := templ.Attributes{}
		if i != 0 {
			attrs["hx-swap-oob"] = "true"
		}
		switch componentSpecifiers[i] {
		case models.NewChoreTypeComponentBaseInputs:
			components = append(components, baseInputs(attrs))
		case models.NewChoreTypeComponentInputTypeSelector:
			components = append(components, selectInputTypeButtonGroup(attrs))
		default:
			logging.Default().Error("unknown new chore type component", "component", componentSpecifiers[i])
		}
	}
	components = append(components, submitRow(templ.Attributes{"hx-swap-oob": "true"}))

	return templ.Join(components...).Render(ctx, w)
}

func isManual(ctx context.Context) bool {
	return models.GetIsManual(ctx)
}

func selectedChoreType(ctx context.Context) *models.ChoreType {
	return models.GetSelectedChoreType(ctx)
}
