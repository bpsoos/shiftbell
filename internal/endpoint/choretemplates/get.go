package choretemplates

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/labstack/echo/v5"
)

func (h *Handler) Get(ctx *echo.Context) error {
	if !supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}

	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: "invalid chore template id"},
		)
	}

	details, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return h.renderError(
				ctx,
				http.StatusNotFound,
				errorResponse{
					Error:   models.ErrNotFound.Error(),
					Links:   collectionLink(),
					Actions: api.Relations{},
				},
			)
		}
		logging.Default().Error("get chore template", "err", err)
		return h.renderError(
			ctx,
			http.StatusInternalServerError,
			errorResponse{Error: "something went wrong", Links: collectionLink()},
		)
	}

	return h.renderDetail(ctx, http.StatusOK, newRepresentation(&details.ChoreTemplate))
}
