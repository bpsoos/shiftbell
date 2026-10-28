package schedules

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/binding"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/schedules"
	validationerrors "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/schedules"
	"github.com/labstack/echo/v5"
)

func (h *Handler) renderEditedSchedule(
	ctx *echo.Context,
	schedule *models.Schedule,
) error {
	if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML {
		ctx.Response().Header().Set("HX-Replace-Url", resourceHref(schedule.Id))
	}
	return h.renderDetail(
		ctx,
		http.StatusOK,
		newRepresentation(schedule),
		"Schedule updated.",
	)
}

func (h *Handler) renderEditRequestError(
	ctx *echo.Context,
	id int,
	request editRequest,
	err error,
) error {
	if errors.Is(err, validationerrors.ErrInvalidInterval) {
		return h.renderEditFailure(ctx, id, request, err)
	}
	if errors.Is(err, binding.ErrUnsupportedMediaType) {
		return h.renderError(
			ctx,
			http.StatusUnsupportedMediaType,
			editErrorResponse(id, err.Error()),
		)
	}
	return h.renderError(
		ctx,
		http.StatusBadRequest,
		editErrorResponse(id, "invalid request body"),
	)
}

func (h *Handler) renderEditFailure(
	ctx *echo.Context,
	id int,
	request editRequest,
	err error,
) error {
	if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML {
		return h.renderEditFormError(ctx, id, request.form(id), err)
	}
	return h.renderEditError(ctx, id, err)
}

func (request editRequest) form(id int) viewmodels.EditForm {
	return viewmodels.EditForm{
		ActionHref: resourceHref(id), CancelHref: resourceHref(id),
		Name: request.Name, IntervalDays: request.intervalValue,
	}
}

func (h *Handler) renderEditFormError(
	ctx *echo.Context,
	id int,
	form viewmodels.EditForm,
	err error,
) error {
	status := http.StatusUnprocessableEntity
	switch {
	case errors.Is(err, models.ErrNameConflict):
		status = http.StatusConflict
		form.NameError = "An active schedule already uses this name."
	case errors.Is(err, validationerrors.ErrInvalidName):
		form.NameError = "Enter a name between 1 and 200 characters."
	case errors.Is(err, validationerrors.ErrInvalidInterval):
		form.IntervalDaysError = "Enter a whole number between 1 and 3650."
	case errors.Is(err, models.ErrInactive), errors.Is(err, models.ErrNotFound):
		return h.renderEditError(ctx, id, err)
	default:
		logging.Default().Error("edit schedule form", "err", err)
		status = http.StatusInternalServerError
		form.SummaryError = "The schedule could not be updated. Try again."
	}
	return h.renderEditForm(ctx, status, form)
}
