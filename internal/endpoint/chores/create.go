package chores

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/bpsoos/shiftbell/internal/endpoint/binding"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	api "github.com/bpsoos/shiftbell/internal/models/api"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	choretemplatemodels "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	"github.com/labstack/echo/v5"
)

type createChoreRequest struct {
	Name                string `form:"name"              json:"name"`
	Description         string `form:"description"       json:"description"`
	Deadline            string `form:"deadline"          json:"deadline"`
	ChoreTemplateId     *int   `form:"chore_template_id" json:"chore_template_id,omitempty"`
	ScheduleName        string `form:"schedule_name"     json:"schedule_name,omitempty"`
	IntervalDays        *int   `form:"interval_days"     json:"interval_days,omitempty"`
	SaveAsChoreTemplate bool   `                         json:"save_as_chore_template"`
}

func (h *Handler) create(ctx *echo.Context) error {
	var request createChoreRequest
	if err := binding.Bind(ctx, &request); err != nil {
		if errors.Is(err, binding.ErrUnsupportedMediaType) {
			return h.renderError(
				ctx,
				http.StatusUnsupportedMediaType,
				apiErrorResponse{Error: binding.ErrUnsupportedMediaType.Error()},
			)
		}
		return h.renderCreateFormError(
			ctx,
			http.StatusBadRequest,
			request,
			formFeedback{Error: apiErrorResponse{Error: "invalid JSON"}},
		)
	}
	if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML {
		request.SaveAsChoreTemplate = false
	}
	if request.IntervalDays != nil {
		return h.scheduledRecurrenceNotImplemented(ctx)
	}
	deadline, err := time.Parse(time.DateOnly, request.Deadline)
	if err != nil {
		response := apiErrorResponse{Error: validationerrors.ErrInvalidDeadline.Error()}
		return h.renderCreateFormError(
			ctx,
			http.StatusUnprocessableEntity,
			request,
			formFeedback{
				Values:      createFormValues(request),
				FieldErrors: map[string]string{"deadline": response.Error},
				Error:       response,
			},
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
		if errors.Is(err, choretemplatemodels.ErrNameConflict) {
			response := apiErrorResponse{
				Error: choretemplatemodels.ErrNameConflict.Error(),
				Links: api.Relations{},
				Actions: api.Relations{
					{Rel: "create", Href: "/chores"},
				},
			}
			return h.renderCreateFormError(
				ctx,
				http.StatusConflict,
				request,
				formFeedback{
					Values:      createFormValues(request),
					FieldErrors: map[string]string{"name": response.Error},
					Error:       response,
				},
			)
		}
		if errors.Is(err, choretemplatemodels.ErrInactive) {
			response := apiErrorResponse{
				Error: choretemplatemodels.ErrInactive.Error(),
				Links: api.Relations{
					{Rel: "collection", Href: "/chores"},
				},
				Actions: api.Relations{
					{Rel: "create", Href: "/chores"},
				},
			}
			return h.renderCreateFormError(
				ctx,
				http.StatusUnprocessableEntity,
				request,
				formFeedback{
					Values: createFormValues(request),
					Error:  response,
				},
			)
		}
		if isInvalidChoreCreateRequest(err) {
			response := apiErrorResponse{Error: err.Error()}
			return h.renderCreateFormError(
				ctx,
				http.StatusUnprocessableEntity,
				request,
				formFeedback{
					Values:      createFormValues(request),
					FieldErrors: createFieldErrors(err),
					Error:       response,
				},
			)
		}
		logging.Default().Error("create chore", "err", err)
		return h.renderError(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}
	if result == nil || result.Chore == nil {
		logging.Default().Error("create chore returned no chore")
		return h.renderError(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}

	response := newChoreResponse(result.Chore)
	return h.renderCreated(ctx, choreRepresentation{
		Response: response,
		Actions:  activeOneOffActions(response.Links.Href("self")),
	})
}

func createFormValues(request createChoreRequest) map[string]string {
	values := map[string]string{
		"name":          request.Name,
		"description":   request.Description,
		"deadline":      request.Deadline,
		"schedule_name": request.ScheduleName,
	}
	if request.ChoreTemplateId != nil {
		values["chore_template_id"] = strconv.Itoa(*request.ChoreTemplateId)
	}
	return values
}

func createFieldErrors(err error) map[string]string {
	field := ""
	switch {
	case errors.Is(err, validationerrors.ErrInvalidName):
		field = "name"
	case errors.Is(err, validationerrors.ErrInvalidDescription):
		field = "description"
	case errors.Is(err, validationerrors.ErrInvalidDeadline):
		field = "deadline"
	}
	if field == "" {
		return nil
	}
	return map[string]string{field: err.Error()}
}

func isInvalidChoreCreateRequest(err error) bool {
	return errors.Is(err, validationerrors.ErrInvalidName) ||
		errors.Is(err, validationerrors.ErrInvalidDescription) ||
		errors.Is(err, validationerrors.ErrInvalidDeadline) ||
		errors.Is(err, validationerrors.ErrInvalidInterval)
}
