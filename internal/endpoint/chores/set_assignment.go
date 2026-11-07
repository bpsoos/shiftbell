package chores

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/binding"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/labstack/echo/v5"
)

func (h *Handler) SetAssignment(ctx *echo.Context) error {
	if !supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	id, err := parseChoreID(ctx)
	if err != nil {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: err.Error()},
		)
	}
	owner, err := parseAssignmentRequest(ctx)
	if err != nil {
		if errors.Is(err, models.ErrInvalidOwnerProfile) {
			return h.renderAssignmentError(ctx, err)
		}
		status := http.StatusBadRequest
		message := "invalid request body"
		if errors.Is(err, binding.ErrUnsupportedMediaType) {
			status = http.StatusUnsupportedMediaType
			message = err.Error()
		}
		return h.renderError(
			ctx,
			status,
			apiErrorResponse{Error: message},
		)
	}
	chore, err := h.service.SetAssignment(
		ctx.Request().Context(),
		&models.SetAssignmentParams{
			Id:             id,
			OwnerProfileId: owner,
		},
	)
	if err != nil {
		return h.renderAssignmentError(ctx, err)
	}
	return h.renderAssignment(ctx, newChoreRepresentation(chore))
}

func (h *Handler) renderAssignmentError(
	ctx *echo.Context,
	err error,
) error {
	if errors.Is(err, models.ErrInvalidOwnerProfile) {
		return h.renderError(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: models.ErrInvalidOwnerProfile.Error()},
		)
	}
	if errors.Is(err, models.ErrNotFound) {
		return h.renderError(
			ctx,
			http.StatusNotFound,
			apiErrorResponse{Error: models.ErrNotFound.Error()},
		)
	}
	logging.Default().Error("set chore assignment", "err", err)
	return h.renderError(
		ctx,
		http.StatusInternalServerError,
		apiErrorResponse{Error: "something went wrong"},
	)
}
