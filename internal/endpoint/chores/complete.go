package chores

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/bpsoos/shiftbell/internal/endpoint/binding"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

type completeChoreRequest struct {
	CompletedOn string `form:"completed_on" json:"completed_on"`
}

func (h *Handler) Complete(ctx *echo.Context) error {
	if hypermedia.Accepts(ctx.Request()) {
		return h.completeVendorJSON(ctx)
	}
	if !acceptsHTMXHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	return h.completeHTMX(ctx)
}

func (h *Handler) completeVendorJSON(ctx *echo.Context) error {
	id, err := parseChoreID(ctx)
	if err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: err.Error()},
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
	result, err := h.service.Complete(
		ctx.Request().Context(),
		&choremodels.CompleteChoreParams{Id: id, CompletedOn: completedOn},
	)
	if err != nil {
		return renderVendorJSONCompletionError(ctx, err)
	}
	if result == nil || result.Chore == nil {
		logging.Default().Error("complete chore returned no chore")
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}
	return hypermedia.JSON(ctx, http.StatusOK, newChoreRepresentation(result.Chore))
}

func (h *Handler) completeHTMX(ctx *echo.Context) error {
	id, err := parseChoreID(ctx)
	if err != nil {
		return hypermedia.NoContent(ctx, http.StatusUnprocessableEntity)
	}
	var request completeChoreRequest
	if err := binding.Bind(ctx, &request); err != nil {
		if errors.Is(err, binding.ErrUnsupportedMediaType) {
			return hypermedia.NoContent(ctx, http.StatusUnsupportedMediaType)
		}
		return hypermedia.NoContent(ctx, http.StatusBadRequest)
	}
	completedOn, err := time.Parse(time.DateOnly, request.CompletedOn)
	if err != nil {
		return h.renderCompletionDateError(ctx, id, request.CompletedOn)
	}
	result, err := h.service.Complete(
		ctx.Request().Context(),
		&choremodels.CompleteChoreParams{Id: id, CompletedOn: completedOn},
	)
	if err != nil {
		return h.renderHTMXCompletionError(ctx, id, request.CompletedOn, err)
	}
	if result == nil || result.Chore == nil {
		logging.Default().Error("complete chore returned no chore")
		return hypermedia.NoContent(ctx, http.StatusInternalServerError)
	}
	setFlashCookie(ctx, choreCompletedFlashValue)
	ctx.Response().Header().Set("HX-Trigger", "choreCompleted")
	return hypermedia.NoContent(ctx, http.StatusNoContent)
}

func renderVendorJSONCompletionError(ctx *echo.Context, err error) error {
	if errors.Is(err, validationerrors.ErrInvalidCompletionDate) {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: validationerrors.ErrInvalidCompletionDate.Error()},
		)
	}
	if errors.Is(err, choremodels.ErrNotFound) {
		return hypermedia.JSON(
			ctx,
			http.StatusNotFound,
			apiErrorResponse{Error: choremodels.ErrNotFound.Error()},
		)
	}
	logging.Default().Error("complete chore", "err", err)
	return hypermedia.JSON(
		ctx,
		http.StatusInternalServerError,
		apiErrorResponse{Error: "something went wrong"},
	)
}

func (h *Handler) renderHTMXCompletionError(
	ctx *echo.Context,
	id int,
	completedOn string,
	err error,
) error {
	if errors.Is(err, validationerrors.ErrInvalidCompletionDate) {
		return h.renderCompletionDateError(ctx, id, completedOn)
	}
	if errors.Is(err, choremodels.ErrNotFound) {
		return hypermedia.NoContent(ctx, http.StatusNotFound)
	}
	logging.Default().Error("complete chore", "err", err)
	return hypermedia.NoContent(ctx, http.StatusInternalServerError)
}
