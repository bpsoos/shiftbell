package chores

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
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
	chores *choremodels.GetChoreBatchResult,
) error {
	return table(offset, limit, chores).Render(ctx, w)
}

func (t *Templater) Page(
	ctx context.Context,
	w io.Writer,
	offset int,
	limit int,
	chores *choremodels.GetChoreBatchResult,
	choreTemplates *choretemplatemodels.GetChoreTemplateBatchResult,
	selectedChoreTemplate *choretemplatemodels.ChoreTemplate,
) error {
	return page(offset, limit, chores, choreTemplates, selectedChoreTemplate).Render(ctx, w)
}

func (t *Templater) Chore(
	ctx context.Context,
	w io.Writer,
	chore *choremodels.Chore,
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
	chores *choremodels.GetChoreBatchResult,
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
	componentSpecifiers ...choremodels.NewChoreTemplateComponent,
) error {
	components := make([]templ.Component, 0)
	for i := range componentSpecifiers {
		attrs := templ.Attributes{}
		if i != 0 {
			attrs["hx-swap-oob"] = "true"
		}
		switch componentSpecifiers[i] {
		case choremodels.NewChoreTemplateComponentBaseInputs:
			components = append(components, baseInputs(attrs))
		case choremodels.NewChoreTemplateComponentInputTypeSelector:
			components = append(components, selectInputTypeButtonGroup(attrs))
		default:
			logging.Default().Error("unknown new chore template component", "component", componentSpecifiers[i])
		}
	}
	components = append(components, submitRow(templ.Attributes{"hx-swap-oob": "true"}))

	return templ.Join(components...).Render(ctx, w)
}

func isManual(ctx context.Context) bool {
	return choremodels.GetIsManual(ctx)
}

func selectedChoreTemplate(ctx context.Context) *choretemplatemodels.ChoreTemplate {
	return choremodels.GetSelectedChoreTemplate(ctx)
}
