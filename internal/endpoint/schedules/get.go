package schedules

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/labstack/echo/v5"
)

func (h *Handler) Get(ctx *echo.Context) error {
	id, err := parseScheduleID(ctx)
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: err.Error(), Links: collectionLink()},
		)
	}
	schedule, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.renderGetError(ctx, err)
	}
	return h.renderDetail(ctx, http.StatusOK, newRepresentation(schedule), "")
}

func (h *Handler) renderGetError(ctx *echo.Context, err error) error {
	if errors.Is(err, models.ErrNotFound) {
		return h.renderError(
			ctx,
			http.StatusNotFound,
			errorResponse{Error: models.ErrNotFound.Error(), Links: collectionLink()},
		)
	}
	logging.Default().Error("get schedule", "err", err)
	return h.renderError(
		ctx,
		http.StatusInternalServerError,
		errorResponse{Error: "something went wrong", Links: collectionLink()},
	)
}
