package chores

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/labstack/echo/v5"
)

func (h *Handler) editForm(ctx *echo.Context) error {
	id, err := parseChoreID(ctx)
	if err != nil {
		return h.renderError(ctx, http.StatusUnprocessableEntity, apiErrorResponse{
			Error: err.Error(),
			Links: api.Relations{{Rel: "collection", Href: choreCollectionHref}},
		})
	}

	chore, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.renderEditLoadError(ctx, err)
	}
	if message := nonEditableMessage(chore); message != "" {
		return h.renderNonEditableChoreError(ctx, chore, message)
	}
	return h.renderEditForm(ctx, http.StatusOK, editFormViewModel(chore))
}

func (h *Handler) renderEditLoadError(ctx *echo.Context, err error) error {
	if errors.Is(err, choremodels.ErrNotFound) {
		return h.renderError(ctx, http.StatusNotFound, apiErrorResponse{
			Error: choremodels.ErrNotFound.Error(),
			Links: api.Relations{{Rel: "collection", Href: choreCollectionHref}},
		})
	}
	logging.Default().Error("load chore for editing", "err", err)
	return h.renderError(ctx, http.StatusInternalServerError, apiErrorResponse{
		Error: "something went wrong",
		Links: api.Relations{{Rel: "collection", Href: choreCollectionHref}},
	})
}
