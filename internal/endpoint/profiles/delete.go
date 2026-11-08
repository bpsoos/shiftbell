package profiles

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	"github.com/labstack/echo/v5"
)

func (h *Handler) Delete(ctx *echo.Context) error {
	if !hypermedia.Accepts(ctx.Request()) {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		return hypermedia.JSON(ctx, http.StatusUnprocessableEntity, api.ErrorResponse{
			Error: "invalid profile id", Links: collectionLink(),
		})
	}
	if err := h.service.Delete(
		ctx.Request().Context(),
		id,
	); err != nil &&
		!errors.Is(err, models.ErrNotFound) {
		return h.internalError(ctx, "delete profile", err)
	}
	return hypermedia.NoContent(ctx, http.StatusNoContent)
}
