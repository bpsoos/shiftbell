package chores

import (
	"errors"
	"mime"
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
	Name                string `form:"name"                   json:"name"`
	Description         string `form:"description"            json:"description"`
	Deadline            string `form:"deadline"               json:"deadline"`
	ChoreTemplateId     *int   `form:"chore_template_id"      json:"chore_template_id,omitempty"`
	ScheduleName        string `form:"schedule_name"          json:"schedule_name,omitempty"`
	IntervalDays        *int   `form:"interval_days"          json:"interval_days,omitempty"`
	SaveAsChoreTemplate bool   `form:"save_as_chore_template" json:"save_as_chore_template"`
}

func (h *Handler) create(ctx *echo.Context) error {
	var request createChoreRequest
	if err := binding.Bind(ctx, &request); err != nil {
		return h.renderCreateBindingError(ctx, request, err)
	}
	request.SaveAsChoreTemplate = saveAsTemplateRequested(ctx, request)
	if request.IntervalDays != nil {
		return h.scheduledRecurrenceNotImplemented(ctx)
	}
	deadline, err := time.Parse(time.DateOnly, request.Deadline)
	if err != nil {
		return h.renderInvalidCreateDeadline(ctx, request)
	}

	result, err := h.service.Create(
		ctx.Request().Context(),
		request.params(deadline),
	)
	if err != nil {
		return h.renderCreateError(ctx, request, err)
	}
	if result == nil || result.Chore == nil {
		return h.renderMissingCreateResult(ctx)
	}
	setCreateSuccessFlash(ctx, request)

	return h.renderCreated(ctx, newChoreRepresentation(result.Chore))
}

func saveAsTemplateRequested(
	ctx *echo.Context,
	request createChoreRequest,
) bool {
	requestMediaType, _, _ := mime.ParseMediaType(
		ctx.Request().Header.Get(echo.HeaderContentType),
	)
	if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML &&
		requestMediaType != echo.MIMEApplicationForm &&
		requestMediaType != echo.MIMEMultipartForm {
		return false
	}
	return request.SaveAsChoreTemplate
}

func (request createChoreRequest) params(
	deadline time.Time,
) *choremodels.CreateChoreParams {
	return &choremodels.CreateChoreParams{
		Name:                request.Name,
		Description:         request.Description,
		Deadline:            deadline,
		ChoreTemplateId:     request.ChoreTemplateId,
		ScheduleName:        request.ScheduleName,
		IntervalDays:        request.IntervalDays,
		SaveAsChoreTemplate: request.SaveAsChoreTemplate,
	}
}

func (h *Handler) renderCreateBindingError(
	ctx *echo.Context,
	request createChoreRequest,
	err error,
) error {
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

func (h *Handler) renderInvalidCreateDeadline(
	ctx *echo.Context,
	request createChoreRequest,
) error {
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

func (h *Handler) renderCreateError(
	ctx *echo.Context,
	request createChoreRequest,
	err error,
) error {
	if errors.Is(err, choretemplatemodels.ErrNameConflict) {
		return h.renderCreateNameConflict(ctx, request)
	}
	if errors.Is(err, choretemplatemodels.ErrInactive) {
		return h.renderInactiveTemplateCreateError(ctx, request)
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

func (h *Handler) renderCreateNameConflict(
	ctx *echo.Context,
	request createChoreRequest,
) error {
	response := apiErrorResponse{
		Error:   choretemplatemodels.ErrNameConflict.Error(),
		Links:   api.Relations{},
		Actions: api.Relations{createChoreSubmissionAction()},
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

func (h *Handler) renderInactiveTemplateCreateError(
	ctx *echo.Context,
	request createChoreRequest,
) error {
	response := apiErrorResponse{
		Error: choretemplatemodels.ErrInactive.Error(),
		Links: api.Relations{
			{Rel: "collection", Href: choreCollectionHref},
		},
		Actions: api.Relations{createChoreSubmissionAction()},
	}
	return h.renderCreateFormError(
		ctx,
		http.StatusUnprocessableEntity,
		request,
		formFeedback{Values: createFormValues(request), Error: response},
	)
}

func (h *Handler) renderMissingCreateResult(ctx *echo.Context) error {
	logging.Default().Error("create chore returned no chore")
	return h.renderError(
		ctx,
		http.StatusInternalServerError,
		apiErrorResponse{Error: "something went wrong"},
	)
}

func setCreateSuccessFlash(ctx *echo.Context, request createChoreRequest) {
	if request.SaveAsChoreTemplate &&
		hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML {
		setFlashCookie(ctx, choreAndTemplateCreatedFlashValue)
	}
}

func createFormValues(request createChoreRequest) map[string]string {
	values := map[string]string{
		"name":                   request.Name,
		"description":            request.Description,
		"deadline":               request.Deadline,
		"schedule_name":          request.ScheduleName,
		"save_as_chore_template": strconv.FormatBool(request.SaveAsChoreTemplate),
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
