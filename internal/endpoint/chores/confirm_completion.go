package chores

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	"github.com/labstack/echo/v5"
)

func (h *Handler) ConfirmCompletion(ctx *echo.Context) error {
	ctx.Response().Header().Set(echo.HeaderVary, "Accept, HX-Request")
	if !acceptsHTMXHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return hypermedia.NoContent(ctx, http.StatusUnprocessableEntity)
	}
	chore, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return renderCompletionLoadError(ctx, err)
	}
	return h.renderCompletionDialog(
		ctx,
		http.StatusOK,
		completionDialogViewModel(newChoreResponse(chore)),
	)
}

func (h *Handler) renderCompletionDateError(
	ctx *echo.Context,
	id int,
	completedOn string,
) error {
	chore, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return renderCompletionLoadError(ctx, err)
	}
	model := completionDialogViewModel(newChoreResponse(chore))
	model.CompletedOn = choreviewmodels.Field{
		Value: completedOn,
		Error: validationerrors.ErrInvalidCompletionDate.Error(),
	}
	return h.renderCompletionDialog(ctx, http.StatusUnprocessableEntity, model)
}

func renderCompletionLoadError(ctx *echo.Context, err error) error {
	if errors.Is(err, choremodels.ErrNotFound) {
		return hypermedia.NoContent(ctx, http.StatusNotFound)
	}
	logging.Default().Error("load chore for completion", "err", err)
	return hypermedia.NoContent(ctx, http.StatusInternalServerError)
}
