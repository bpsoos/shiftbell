package chores

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

func (h *Handler) CorrectCompletion(ctx *echo.Context) error {
	id, err := parseChoreID(ctx)
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: err.Error()},
		)
	}
	var request completeChoreRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&request); err != nil {
		return h.renderError(
			ctx,
			http.StatusBadRequest,
			apiErrorResponse{Error: "invalid JSON"},
		)
	}
	completedOn, err := time.Parse(time.DateOnly, request.CompletedOn)
	if err != nil {
		return h.renderError(
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
		return h.renderCorrectionError(ctx, err)
	}
	if corrected == nil {
		logging.Default().Error("correct chore completion returned no chore")
		return h.renderError(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}
	return h.renderDetail(
		ctx,
		http.StatusOK,
		newChoreRepresentation(corrected),
		"Completion date corrected.",
	)
}

func (h *Handler) renderCorrectionError(ctx *echo.Context, err error) error {
	if errors.Is(err, validationerrors.ErrInvalidCompletionDate) {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: validationerrors.ErrInvalidCompletionDate.Error()},
		)
	}
	if errors.Is(err, choremodels.ErrNotFound) {
		return h.renderError(
			ctx,
			http.StatusNotFound,
			apiErrorResponse{Error: choremodels.ErrNotFound.Error()},
		)
	}
	logging.Default().Error("correct chore completion", "err", err)
	return h.renderError(
		ctx,
		http.StatusInternalServerError,
		apiErrorResponse{Error: "something went wrong"},
	)
}
