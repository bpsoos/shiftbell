package choretemplates

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	models "github.com/bpsoos/shiftbell/internal/models/choretemplates"
	"github.com/labstack/echo/v5"
)

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
	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			errorResponse{Error: "invalid chore template id"},
		)
	}
	deactivated, err := h.service.Deactivate(ctx.Request().Context(), id)
	if err != nil {
		var referencesError *models.ActiveScheduleReferencesError
		if errors.As(err, &referencesError) {
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
	return hypermedia.JSON(ctx, http.StatusOK, newRepresentation(deactivated))
}

func (h *Handler) deactivateHTMX(ctx *echo.Context) error {
	id, err := strconv.Atoi(ctx.ParamOr("id", ""))
	if err != nil || id <= 0 {
		return hypermedia.NoContent(ctx, http.StatusUnprocessableEntity)
	}
	deactivated, err := h.service.Deactivate(ctx.Request().Context(), id)
	if err != nil {
		var referencesError *models.ActiveScheduleReferencesError
		switch {
		case errors.As(err, &referencesError):
			return h.renderDeactivationDialogError(
				ctx,
				http.StatusConflict,
				id,
				"This template cannot be deactivated while active schedules use it.",
			)
		case errors.Is(err, models.ErrInactive):
			return h.renderInactiveDeactivation(ctx, id)
		case errors.Is(err, models.ErrNotFound):
			return hypermedia.NoContent(ctx, http.StatusNotFound)
		default:
			logging.Default().Error("deactivate chore template", "err", err)
			return h.renderDeactivationDialogError(
				ctx,
				http.StatusInternalServerError,
				id,
				"The template could not be deactivated. Try again.",
			)
		}
	}
	if deactivated == nil {
		logging.Default().Error("deactivate chore template returned no template")
		return h.renderDeactivationDialogError(
			ctx,
			http.StatusInternalServerError,
			id,
			"The template could not be deactivated. Try again.",
		)
	}

	setFlashCookie(ctx, templateDeactivatedFlashValue)
	ctx.Response().Header().Set("HX-Trigger", "templateDeactivated")
	return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, "/chore-templates")
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
