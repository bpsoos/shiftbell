package chores

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	choreviewmodels "github.com/bpsoos/shiftbell/internal/models/view/chores"
	confirmationviewmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	"github.com/labstack/echo/v5"
)

func (h *Handler) ConfirmCompletion(ctx *echo.Context) error {
	ctx.Response().Header().Set(echo.HeaderVary, "Accept, HX-Request")
	if !acceptsHTMXHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := parseChoreID(ctx)
	if err != nil {
		return hypermedia.NoContent(ctx, http.StatusUnprocessableEntity)
	}
	chore, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.renderCompletionLoadError(ctx, err)
	}
	if chore.Status != choremodels.ChoreStatusActive {
		return h.renderConfirmationDialog(
			ctx,
			http.StatusUnprocessableEntity,
			confirmationviewmodels.Dialog{
				Heading: "Chore already completed",
				Supporting: []string{
					"This chore is already complete. Close this dialog and refresh the chores.",
				},
			},
		)
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
		return h.renderCompletionLoadError(ctx, err)
	}
	model := completionDialogViewModel(newChoreResponse(chore))
	model.CompletedOn = choreviewmodels.Field{
		Value: completedOn,
		Error: validationerrors.ErrInvalidCompletionDate.Error(),
	}
	return h.renderCompletionDialog(ctx, http.StatusUnprocessableEntity, model)
}

func (h *Handler) renderCompletionLoadError(ctx *echo.Context, err error) error {
	status := http.StatusInternalServerError
	message := "The chore could not be loaded. Close this dialog and try again."
	if errors.Is(err, choremodels.ErrNotFound) {
		status = http.StatusNotFound
		message = "This chore could not be found. Close this dialog and refresh the chores."
	} else {
		logging.Default().Error("load chore for completion", "err", err)
	}
	return h.renderConfirmationDialog(ctx, status, confirmationviewmodels.Dialog{
		Heading: "Unable to complete chore", Error: message,
	})
}
