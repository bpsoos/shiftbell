package schedules

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

func (h *Handler) Edit(ctx *echo.Context) error {
	if !supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := parseScheduleID(ctx)
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: err.Error(), Links: collectionLink()},
		)
	}
	request, err := parseEditRequest(ctx)
	if err != nil {
		return h.renderEditRequestError(ctx, id, request, err)
	}
	schedule, err := h.service.Edit(ctx.Request().Context(), &models.EditScheduleParams{
		Id: id, Name: request.Name, IntervalDays: request.IntervalDays,
	})
	if err != nil {
		return h.renderEditFailure(ctx, id, request, err)
	}
	return h.renderEditedSchedule(ctx, schedule)
}

func (h *Handler) renderEditError(ctx *echo.Context, id int, err error) error {
	switch {
	case errors.Is(err, models.ErrNameConflict):
		return h.renderError(
			ctx,
			http.StatusConflict,
			editErrorResponse(id, models.ErrNameConflict.Error()),
		)
	case errors.Is(err, validationerrors.ErrInvalidName):
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			editErrorResponse(id, validationerrors.ErrInvalidName.Error()),
		)
	case errors.Is(err, validationerrors.ErrInvalidInterval):
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			editErrorResponse(id, validationerrors.ErrInvalidInterval.Error()),
		)
	case errors.Is(err, models.ErrInactive):
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: models.ErrInactive.Error(), Links: collectionLink()},
		)
	case errors.Is(err, models.ErrNotFound):
		return h.renderError(
			ctx,
			http.StatusNotFound,
			errorResponse{Error: models.ErrNotFound.Error(), Links: collectionLink()},
		)
	default:
		logging.Default().Error("edit schedule", "err", err)
		return h.renderError(
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
