package chores

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/labstack/echo/v5"
)

func (h *Handler) get(ctx *echo.Context) error {
	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: "invalid chore id"},
		)
	}
	chore, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, choremodels.ErrNotFound) {
			return h.renderError(
				ctx,
				http.StatusNotFound,
				apiErrorResponse{
					Error: choremodels.ErrNotFound.Error(),
					Links: map[string]api.Link{
						"collection": {Href: "/chores"},
					},
					Actions: map[string]api.Action{},
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

	response := newChoreResponse(chore)
	actions := activeOneOffActions(response.Links["self"].Href)
	if chore.Status == choremodels.ChoreStatusCompleted {
		actions = completedOneOffActions(response.Links["self"].Href)
	}
	return h.renderDetail(ctx, http.StatusOK, choreRepresentation{
		Response: response,
		Actions:  actions,
	})
}
