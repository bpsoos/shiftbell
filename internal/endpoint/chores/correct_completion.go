package chores

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

func (h *Handler) CorrectCompletion(ctx *echo.Context) error {
	if !hypermedia.Accepts(ctx.Request()) {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: "invalid chore id"},
		)
	}
	var request completeChoreRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&request); err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusBadRequest,
			apiErrorResponse{Error: "invalid JSON"},
		)
	}
	completedOn, err := time.Parse(time.DateOnly, request.CompletedOn)
	if err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: validationerrors.ErrInvalidCompletionDate.Error()},
		)
	}
	corrected, err := h.service.CorrectCompletion(
		ctx.Request().Context(),
		&choremodels.CorrectCompletionParams{Id: id, CompletedOn: completedOn},
	)
	if err != nil {
		if errors.Is(err, validationerrors.ErrInvalidCompletionDate) {
			return hypermedia.JSON(
				ctx,
				http.StatusUnprocessableEntity,
				apiErrorResponse{
					Error: validationerrors.ErrInvalidCompletionDate.Error(),
				},
			)
		}
		if errors.Is(err, choremodels.ErrNotFound) {
			return hypermedia.JSON(
				ctx,
				http.StatusNotFound,
				apiErrorResponse{Error: choremodels.ErrNotFound.Error()},
			)
		}
		logging.Default().Error("correct chore completion", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}
	if corrected == nil {
		logging.Default().Error("correct chore completion returned no chore")
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}
	response := newChoreResponse(corrected)
	return hypermedia.JSON(ctx, http.StatusOK, choreRepresentation{
		Response: response,
		Actions:  completedOneOffActions(response.Links["self"].Href),
	})
}
