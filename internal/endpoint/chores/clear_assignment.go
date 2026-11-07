package chores

import (
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/labstack/echo/v5"
)

func (h *Handler) ClearAssignment(ctx *echo.Context) error {
	if !supported(ctx) {
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
	chore, err := h.service.ClearAssignment(ctx.Request().Context(), id)
	if err != nil {
		return h.renderAssignmentError(ctx, err)
	}
	return h.renderAssignment(ctx, newChoreRepresentation(chore))
}
