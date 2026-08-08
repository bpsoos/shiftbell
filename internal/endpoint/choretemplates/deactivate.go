package choretemplates

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/labstack/echo/v5"
)

func (h *Handler) Deactivate(ctx *echo.Context) error {
	if !hypermedia.Accepts(ctx.Request()) {
		return hypermedia.NotAcceptable(ctx)
	}

	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: "invalid chore template id"},
		)
	}
	deactivated, err := h.service.Deactivate(ctx.Request().Context(), id)
	if err != nil {
		var referencesError *models.ActiveScheduleReferencesError
		if errors.As(err, &referencesError) {
			return hypermedia.JSON(
				ctx,
				http.StatusConflict,
				resourceErrorResponse(id, referencesError.Error()),
			)
		}
		if errors.Is(err, models.ErrInactive) {
			return hypermedia.JSON(
				ctx,
				http.StatusUnprocessableEntity,
				resourceErrorResponse(id, models.ErrInactive.Error()),
			)
		}
		if errors.Is(err, models.ErrNotFound) {
			return hypermedia.JSON(
				ctx,
				http.StatusNotFound,
				errorResponse{Error: models.ErrNotFound.Error(), Links: collectionLink()},
			)
		}
		logging.Default().Error("deactivate chore template", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			resourceErrorResponse(id, "something went wrong"),
		)
	}
	return hypermedia.JSON(ctx, http.StatusOK, newRepresentation(deactivated))
}
