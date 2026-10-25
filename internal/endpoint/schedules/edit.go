package schedules

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

type editRequest struct {
	Name         string `json:"name"`
	IntervalDays int    `json:"interval_days"`
}

func (h *Handler) Edit(ctx *echo.Context) error {
	if hypermedia.Negotiate(ctx.Request()) != hypermedia.RepresentationJSON {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := parseScheduleID(ctx)
	if err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: err.Error(), Links: collectionLink()},
		)
	}
	var request editRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&request); err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusBadRequest,
			editErrorResponse(id, "invalid request body"),
		)
	}
	schedule, err := h.service.Edit(ctx.Request().Context(), &models.EditScheduleParams{
		Id: id, Name: request.Name, IntervalDays: request.IntervalDays,
	})
	if err != nil {
		return h.renderEditError(ctx, id, err)
	}
	return hypermedia.JSON(ctx, http.StatusOK, newRepresentation(schedule))
}

func (h *Handler) renderEditError(ctx *echo.Context, id int, err error) error {
	switch {
	case errors.Is(err, models.ErrNameConflict):
		return hypermedia.JSON(
			ctx,
			http.StatusConflict,
			editErrorResponse(id, models.ErrNameConflict.Error()),
		)
	case errors.Is(err, validationerrors.ErrInvalidName):
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			editErrorResponse(id, validationerrors.ErrInvalidName.Error()),
		)
	case errors.Is(err, validationerrors.ErrInvalidInterval):
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			editErrorResponse(id, validationerrors.ErrInvalidInterval.Error()),
		)
	case errors.Is(err, models.ErrInactive):
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: models.ErrInactive.Error(), Links: collectionLink()},
		)
	case errors.Is(err, models.ErrNotFound):
		return hypermedia.JSON(
			ctx,
			http.StatusNotFound,
			errorResponse{Error: models.ErrNotFound.Error(), Links: collectionLink()},
		)
	default:
		logging.Default().Error("edit schedule", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			errorResponse{Error: "something went wrong", Links: collectionLink()},
		)
	}
}

func editErrorResponse(id int, message string) errorResponse {
	return errorResponse{
		Error: message, Links: collectionLink(),
		Actions: api.Relations{{Rel: "edit", Href: resourceHref(id)}},
	}
}
