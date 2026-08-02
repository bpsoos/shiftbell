package chores

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/labstack/echo/v5"
)

func (h *Handler) get(ctx *echo.Context) error {
	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: "invalid chore id"},
		)
	}
	chore, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, choremodels.ErrNotFound) {
			return hypermedia.JSON(
				ctx,
				http.StatusNotFound,
				apiErrorResponse{Error: choremodels.ErrNotFound.Error()},
			)
		}
		logging.Default().Error("get chore", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}

	response := newChoreResponse(chore)
	return hypermedia.JSON(ctx, http.StatusOK, choreRepresentation{
		choreResponse: response,
		Actions:       activeOneOffActions(response.Links["self"].Href),
	})
}
