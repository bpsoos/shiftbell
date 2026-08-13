package chores

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	choremodels "github.com/bpsoos/shiftbell/internal/models/chores"
	"github.com/labstack/echo/v5"
)

func (h *Handler) Delete(ctx *echo.Context) error {
	if hypermedia.Accepts(ctx.Request()) {
		return h.deleteVendorJSON(ctx)
	}
	ctx.Response().Header().Set(echo.HeaderVary, "Accept, HX-Request")
	if !acceptsHTMXHTML(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	return h.deleteHTMX(ctx)
}

func (h *Handler) deleteVendorJSON(ctx *echo.Context) error {
	id, err := parseChoreID(ctx)
	if err != nil {
		return hypermedia.JSON(
			ctx,
			http.StatusUnprocessableEntity,
			apiErrorResponse{Error: err.Error()},
		)
	}
	if err := h.service.Delete(ctx.Request().Context(), id); err != nil {
		if errors.Is(err, choremodels.ErrNotFound) {
			return hypermedia.NoContent(ctx, http.StatusNoContent)
		}
		logging.Default().Error("delete chore", "err", err)
		return hypermedia.JSON(
			ctx,
			http.StatusInternalServerError,
			apiErrorResponse{Error: "something went wrong"},
		)
	}
	return hypermedia.NoContent(ctx, http.StatusNoContent)
}

func (h *Handler) deleteHTMX(ctx *echo.Context) error {
	id, err := parseChoreID(ctx)
	if err != nil {
		return hypermedia.NoContent(ctx, http.StatusUnprocessableEntity)
	}
	if err := h.service.Delete(ctx.Request().Context(), id); err != nil {
		if errors.Is(err, choremodels.ErrNotFound) {
			return h.renderDeleted(ctx)
		}
		logging.Default().Error("delete chore", "err", err)
		return h.renderDeletionError(ctx, id)
	}
	return h.renderDeleted(ctx)
}

func (h *Handler) renderDeleted(ctx *echo.Context) error {
	setFlashCookie(ctx, choreDeletedFlashValue)
	ctx.Response().Header().Set("HX-Trigger", "choreDeleted")
	return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, choreCollectionHref)
}

func (h *Handler) renderDeletionError(ctx *echo.Context, id int) error {
	chore, err := h.service.Get(ctx.Request().Context(), id)
	if err != nil {
		return renderDeletionLoadError(ctx, err)
	}
	return h.renderConfirmationDialog(
		ctx,
		http.StatusInternalServerError,
		deletionDialogViewModel(chore, "The chore could not be deleted. Try again."),
	)
}
