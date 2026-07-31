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

func (h *Handler) Get(ctx *echo.Context) error {
	if !hypermedia.Accepts(ctx.Request()) {
		return ctx.NoContent(http.StatusNotAcceptable)
	}

	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return hypermedia.JSON(ctx, http.StatusUnprocessableEntity, errorResponse{Error: "invalid chore template id"})
	}

	details, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return hypermedia.JSON(ctx, http.StatusNotFound, errorResponse{Error: models.ErrNotFound.Error()})
		}
		logging.Default().Error("get chore template", "err", err)
		return hypermedia.JSON(ctx, http.StatusInternalServerError, errorResponse{Error: "something went wrong"})
	}

	return hypermedia.JSON(ctx, http.StatusOK, newResponse(&details.ChoreTemplate))
}
