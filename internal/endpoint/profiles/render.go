package profiles

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	"github.com/bpsoos/shiftbell/internal/logging"
	"github.com/bpsoos/shiftbell/internal/models/api"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	validation "github.com/bpsoos/shiftbell/internal/models/validation"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
)

func (h *Handler) creationFailure(
	ctx *echo.Context,
	form viewmodels.Form,
	err error,
) error {
	status := http.StatusUnprocessableEntity
	var message string
	switch {
	case errors.Is(err, models.ErrNameConflict):
		status = http.StatusConflict
		message = models.ErrNameConflict.Error()
		form.Name.Error = "A profile already uses this name."
	case errors.Is(err, validation.ErrInvalidName):
		message = validation.ErrInvalidName.Error()
		form.Name.Error = "Enter 1–200 printable characters with no whitespace."
	case errors.Is(err, models.ErrAlreadyInitialized):
		status = http.StatusConflict
		message = models.ErrAlreadyInitialized.Error()
		form.SummaryError = "Setup is already complete. Continue to chores."
		form.CancelHref = "/chores"
	default:
		logging.Default().Error("create profile", "err", err)
		status = http.StatusInternalServerError
		message = "The profile could not be created. Try again."
		form.SummaryError = message
	}
	return h.formFailure(ctx, status, form, message)
}

func (h *Handler) formFailure(
	ctx *echo.Context,
	status int,
	form viewmodels.Form,
	message string,
) error {
	if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML {
		return hypermedia.HTML(ctx, status, h.view.Form(form, fullPage(ctx)))
	}
	return hypermedia.JSON(ctx, status, api.ErrorResponse{
		Error: message, Links: collectionLink(),
		Actions: api.Relations{{Rel: "create", Href: form.ActionHref}},
	})
}

func (h *Handler) internalError(ctx *echo.Context, operation string, err error) error {
	logging.Default().Error(operation, "err", err)
	if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML &&
		h.view != nil {
		return hypermedia.HTML(
			ctx,
			http.StatusInternalServerError,
			h.view.Error(viewmodels.Error{
				Message: "Something went wrong. Try again.", CollectionHref: "/profiles",
			}, fullPage(ctx)),
		)
	}
	return hypermedia.JSON(ctx, http.StatusInternalServerError, api.ErrorResponse{
		Error: "something went wrong", Links: collectionLink(),
	})
}
