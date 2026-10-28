package schedules

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	confirmationmodels "github.com/bpsoos/shiftbell/internal/models/view/confirmation"
	"github.com/labstack/echo/v5"
)

func (h *Handler) Deactivate(ctx *echo.Context) error {
	if !supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := parseScheduleID(ctx)
	if err != nil {
		if acceptsHTMXHTML(ctx) {
			return h.renderDeactivationDialog(
				ctx,
				http.StatusUnprocessableEntity,
				confirmationmodels.Dialog{
					Heading: "Unable to deactivate schedule", Error: err.Error(),
				},
			)
		}
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: err.Error(), Links: collectionLink()},
		)
	}
	schedule, err := h.service.Deactivate(ctx.Request().Context(), id)
	if err != nil {
		if acceptsHTMXHTML(ctx) {
			logging.Default().Error("deactivate schedule", "err", err)
			return h.loadDeactivationDialog(
				ctx,
				id,
				http.StatusInternalServerError,
				"The schedule could not be deactivated. Try again.",
			)
		}
		return h.renderDeactivationError(ctx, err)
	}
	if acceptsHTMXHTML(ctx) {
		ctx.Response().Header().Set("HX-Retarget", "#main")
		ctx.Response().Header().Set("HX-Reswap", "outerHTML")
		ctx.Response().Header().Set("HX-Replace-Url", resourceHref(id))
		ctx.Response().Header().Set("HX-Trigger", "scheduleDeactivated")
	}
	return h.renderDetail(
		ctx,
		http.StatusOK,
		newRepresentation(schedule),
		"Schedule deactivated.",
	)
}

func (h *Handler) renderDeactivationError(ctx *echo.Context, err error) error {
	if errors.Is(err, models.ErrNotFound) {
		return h.renderError(
			ctx,
			http.StatusNotFound,
			errorResponse{Error: models.ErrNotFound.Error(), Links: collectionLink()},
		)
	}
	logging.Default().Error("deactivate schedule", "err", err)
	return h.renderError(
		ctx,
		http.StatusInternalServerError,
		errorResponse{Error: "something went wrong", Links: collectionLink()},
	)
}
