package choretemplates

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	schedulemodels "github.com/bpsoos/shiftbell/internal/models/schedules"
	"github.com/labstack/echo/v5"
)

const deactivationFailedMessage = "The template could not be deactivated. Try again."

func (h *Handler) Deactivate(ctx *echo.Context) error {
	if hypermedia.Accepts(ctx.Request()) {
		return h.deactivateVendorJSON(ctx)
	}
	ctx.Response().Header().Set(echo.HeaderVary, "Accept, HX-Request")
	if !acceptsHTMXHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	return h.deactivateHTMX(ctx)
}

func (h *Handler) deactivateVendorJSON(ctx *echo.Context) error {
	id, err := parseChoreTemplateID(ctx)
	if err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: err.Error()},
		)
	}
	deactivated, err := h.service.Deactivate(ctx.Request().Context(), id)
	if err != nil {
		return h.renderVendorJSONDeactivationError(ctx, id, err)
	}
	return hypermedia.JSON(ctx, http.StatusOK, newRepresentation(deactivated))
}

func (h *Handler) deactivateHTMX(ctx *echo.Context) error {
	id, err := parseChoreTemplateID(ctx)
	if err != nil {
		return hypermedia.NoContent(ctx, http.StatusUnprocessableEntity)
	}
	deactivated, err := h.service.Deactivate(ctx.Request().Context(), id)
	if err != nil {
		return h.renderHTMXDeactivationError(ctx, id, err)
	}
	if deactivated == nil {
		return h.renderMissingDeactivationResult(ctx, id)
	}
	return redirectAfterDeactivation(ctx)
}

func (h *Handler) renderVendorJSONDeactivationError(
	ctx *echo.Context,
	id int,
	err error,
) error {
	if referencesError, ok := errors.AsType[*models.ActiveScheduleReferencesError](
		err,
	); ok {
		return hypermedia.JSON(
			ctx,
			http.StatusConflict,
			resourceErrorResponse(id, referencesError.Error()),
		)
	}
	if errors.Is(err, models.ErrInactive) {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			resourceErrorResponse(id, models.ErrInactive.Error()),
		)
	}
	if errors.Is(err, models.ErrNotFound) {
		return hypermedia.JSON(
			ctx,
			http.StatusNotFound,
			errorResponse{Error: models.ErrNotFound.Error(), Links: collectionLink()},
		)
	}
	logging.Default().Error("deactivate chore template", "err", err)
	return hypermedia.JSON(
		ctx,
		http.StatusInternalServerError,
		resourceErrorResponse(id, "something went wrong"),
	)
}

func (h *Handler) renderHTMXDeactivationError(
	ctx *echo.Context,
	id int,
	err error,
) error {
	if references, ok := errors.AsType[*models.ActiveScheduleReferencesError](err); ok {
		return h.renderBlockedDeactivation(
			ctx,
			id,
			references.Schedules,
		)
	}
	if errors.Is(err, models.ErrInactive) {
		return h.renderInactiveDeactivation(ctx, id)
	}
	if errors.Is(err, models.ErrNotFound) {
		return h.renderDeactivationLoadError(ctx, err)
	}
	logging.Default().Error("deactivate chore template", "err", err)
	return h.renderDeactivationDialogError(
		ctx,
		http.StatusInternalServerError,
		id,
		deactivationFailedMessage,
	)
}

func (h *Handler) renderBlockedDeactivation(
	ctx *echo.Context,
	id int,
	schedules []models.ScheduleReference,
) error {
	model := scheduleListViewModel(
		id,
		schedulemodels.ScheduleFilterActive,
		schedules,
		true,
	)
	model.Notice = "Deactivate these active schedules before deactivating the template."
	ctx.Response().Header().Set("HX-Retarget", "#template-schedule-list")
	ctx.Response().Header().Set("HX-Reswap", "innerHTML")
	ctx.Response().Header().Set("HX-Trigger-After-Swap", "templateDeactivationBlocked")
	return h.renderScheduleList(ctx, http.StatusConflict, model)
}

func (h *Handler) renderMissingDeactivationResult(
	ctx *echo.Context,
	id int,
) error {
	logging.Default().Error("deactivate chore template returned no template")
	return h.renderDeactivationDialogError(
		ctx,
		http.StatusInternalServerError,
		id,
		deactivationFailedMessage,
	)
}

func redirectAfterDeactivation(ctx *echo.Context) error {
	setFlashCookie(ctx, templateDeactivatedFlashValue)
	ctx.Response().Header().Set("HX-Trigger", "templateDeactivated")
	return hypermedia.HTMLRedirect(
		ctx,
		http.StatusSeeOther,
		choreTemplateCollectionHref,
	)
}

func (h *Handler) renderDeactivationDialogError(
	ctx *echo.Context,
	status int,
	id int,
	message string,
) error {
	details, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.renderDeactivationLoadError(ctx, err)
	}
	return h.renderConfirmationDialog(
		ctx,
		status,
		deactivationDialogViewModel(&details.ChoreTemplate, message),
	)
}

func (h *Handler) renderInactiveDeactivation(ctx *echo.Context, id int) error {
	details, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return h.renderDeactivationLoadError(ctx, err)
	}
	return h.renderConfirmationDialog(
		ctx,
		http.StatusUnprocessableEntity,
		inactiveDeactivationDialogViewModel(&details.ChoreTemplate),
	)
}
