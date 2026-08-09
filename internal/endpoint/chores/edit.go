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

type editChoreRequest struct {
	Name                    string `json:"name"`
	Description             string `json:"description"`
	Deadline                string `json:"deadline"`
	AlsoUpdateChoreTemplate bool   `json:"also_update_chore_template"`
}

func (h *Handler) edit(ctx *echo.Context) error {
	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: "invalid chore id"},
		)
	}
	var request editChoreRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&request); err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusBadRequest,
			apiErrorResponse{Error: "invalid JSON"},
		)
	}
	deadline, err := time.Parse(time.DateOnly, request.Deadline)
	if err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: validationerrors.ErrInvalidDeadline.Error()},
		)
	}
	existing, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.editChoreError(ctx, err)
	}
	edited, err := h.service.Edit(ctx.Request().Context(), &choremodels.EditChoreParams{
		Id:                      id,
		ScheduleId:              existing.ScheduleId,
		Name:                    request.Name,
		Description:             request.Description,
		Deadline:                deadline,
		AlsoUpdateChoreTemplate: request.AlsoUpdateChoreTemplate,
	})
	if err != nil {
		return h.editChoreError(ctx, err)
	}
	response := newChoreResponse(edited)
	return hypermedia.JSON(ctx, http.StatusOK, choreRepresentation{
		Response: response,
		Actions:  activeOneOffActions(response.Links.Href("self")),
	})
}

func (h *Handler) editChoreError(
	ctx *echo.Context,
	err error,
) error {
	if isEditChoreValidationError(err) {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: err.Error()},
		)
	}
	if errors.Is(err, choremodels.ErrNotFound) {
		return hypermedia.JSON(
			ctx,
			http.StatusNotFound,
			apiErrorResponse{Error: choremodels.ErrNotFound.Error()},
		)
	}
	logging.Default().Error("edit chore", "err", err)
	return hypermedia.JSON(
		ctx,
		http.StatusInternalServerError,
		apiErrorResponse{Error: "something went wrong"},
	)
}

func isEditChoreValidationError(err error) bool {
	return errors.Is(err, validationerrors.ErrInvalidName) ||
		errors.Is(err, validationerrors.ErrInvalidDescription) ||
		errors.Is(err, validationerrors.ErrInvalidDeadline)
}
