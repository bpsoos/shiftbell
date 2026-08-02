package chores

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

type createChoreRequest struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	Deadline            string `json:"deadline"`
	ChoreTemplateId     *int   `json:"chore_template_id,omitempty"`
	ScheduleName        string `json:"schedule_name,omitempty"`
	IntervalDays        *int   `json:"interval_days,omitempty"`
	SaveAsChoreTemplate bool   `json:"save_as_chore_template"`
}

func (h *Handler) create(ctx *echo.Context) error {
	var request createChoreRequest
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

	result, err := h.service.Create(
		ctx.Request().Context(),
		&choremodels.CreateChoreParams{
			Name:                request.Name,
			Description:         request.Description,
			Deadline:            deadline,
			ChoreTemplateId:     request.ChoreTemplateId,
			ScheduleName:        request.ScheduleName,
			IntervalDays:        request.IntervalDays,
			SaveAsChoreTemplate: request.SaveAsChoreTemplate,
		},
	)
	if err != nil {
		if isInvalidChoreCreateRequest(err) {
			return hypermedia.JSON(
				ctx,
				http.StatusUnprocessableEntity,
				apiErrorResponse{Error: err.Error()},
			)
		}
		logging.Default().Error("create chore", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}
	if result == nil || result.Chore == nil {
		logging.Default().Error("create chore returned no chore")
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}

	response := newChoreResponse(result.Chore)
	ctx.Response().Header().Set(echo.HeaderLocation, response.Links["self"].Href)
	return hypermedia.JSON(ctx, http.StatusCreated, choreRepresentation{
		choreResponse: response,
		Actions:       activeOneOffActions(response.Links["self"].Href),
	})
}

func isInvalidChoreCreateRequest(err error) bool {
	return errors.Is(err, validationerrors.ErrInvalidName) ||
		errors.Is(err, validationerrors.ErrInvalidDescription) ||
		errors.Is(err, validationerrors.ErrInvalidDeadline) ||
		errors.Is(err, validationerrors.ErrInvalidInterval)
}
