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

func (h *Handler) EditForm(ctx *echo.Context) error {
	ctx.Response().Header().Set(echo.HeaderVary, "Accept, HX-Request")
	if !acceptsHTMXHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}

	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return h.renderError(ctx, http.StatusUnprocessableEntity, errorResponse{
			Error: "invalid chore template id",
			Links: collectionLink(),
		})
	}

	details, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.renderEditLoadError(ctx, err)
	}
	if details.DeactivatedAt != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			resourceErrorResponse(id, models.ErrInactive.Error()),
		)
	}
	return h.renderEditForm(
		ctx,
		http.StatusOK,
		editFormViewModel(&details.ChoreTemplate),
	)
}

func (h *Handler) renderEditLoadError(ctx *echo.Context, err error) error {
	if errors.Is(err, models.ErrNotFound) {
		return h.renderError(ctx, http.StatusNotFound, errorResponse{
			Error: models.ErrNotFound.Error(),
			Links: collectionLink(),
		})
	}
	logging.Default().Error("load chore template for editing", "err", err)
	return h.renderError(ctx, http.StatusInternalServerError, errorResponse{
		Error: "something went wrong",
		Links: collectionLink(),
	})
}
