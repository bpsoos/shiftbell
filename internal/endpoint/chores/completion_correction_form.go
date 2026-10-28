package chores

import (
	"net/http"
	"time"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
)

func (h *Handler) CompletionCorrectionForm(ctx *echo.Context) error {
	if !acceptsHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := parseChoreID(ctx)
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: err.Error()},
		)
	}
	chore, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.renderGetError(ctx, err)
	}
	if chore.Status != choremodels.ChoreStatusCompleted {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{
				Error: "Only completed chores have a completion date to correct.",
			},
		)
	}
	return h.renderCompletionCorrectionForm(
		ctx,
		http.StatusOK,
		correctionFormModel(id, chore.CompletedOn.Format(time.DateOnly)),
	)
}

func correctionFormModel(id int, completedOn string) viewmodels.CompletionCorrectionForm {
	return viewmodels.CompletionCorrectionForm{
		ActionHref: choreHref(id) + "/completion", CancelHref: choreHref(id),
		CompletedOn: viewmodels.Field{Value: completedOn},
	}
}

func (h *Handler) renderCompletionCorrectionForm(
	ctx *echo.Context,
	status int,
	form viewmodels.CompletionCorrectionForm,
) error {
	return hypermedia.HTML(
		ctx,
		status,
		h.view.CompletionCorrectionForm(form, fullPage(ctx)),
	)
}
