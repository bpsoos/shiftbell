package chores

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/labstack/echo/v5"
)

func (h *Handler) get(ctx *echo.Context) error {
	id, err := parseChoreID(ctx)
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: err.Error()},
		)
	}
	chore, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.renderGetError(ctx, err)
	}

	response := newChoreResponse(chore)
	return h.renderDetail(
		ctx,
		http.StatusOK,
		choreRepresentation{
			Response: response,
			Actions:  actionsForChore(chore),
		},
		"",
	)
}

func (h *Handler) renderGetError(ctx *echo.Context, err error) error {
	if errors.Is(err, choremodels.ErrNotFound) {
		return h.renderError(
			ctx,
			http.StatusNotFound,
			apiErrorResponse{
				Error: choremodels.ErrNotFound.Error(),
				Links: api.Relations{
					{Rel: "collection", Href: choreCollectionHref},
				},
				Actions: api.Relations{},
			},
		)
	}
	logging.Default().Error("get chore", "err", err)
	return h.renderError(
		ctx,
		http.StatusInternalServerError,
		apiErrorResponse{Error: "something went wrong"},
	)
}
