package profiles

import (
	"errors"
	"net/http"

	"github.com/bpsoos/shiftbell/internal/endpoint/binding"
	"github.com/bpsoos/shiftbell/internal/endpoint/hypermedia"
	models "github.com/bpsoos/shiftbell/internal/models/profiles"
	viewmodels "github.com/bpsoos/shiftbell/internal/models/view/profiles"
	"github.com/labstack/echo/v5"
)

type createRequest struct {
	Name string `json:"name" form:"name"`
}

func (h *Handler) Create(ctx *echo.Context) error {
	return h.create(ctx, false)
}

func (h *Handler) create(ctx *echo.Context, setup bool) error {
	if !h.supported(ctx) {
		return hypermedia.NotAcceptable(ctx)
	}
	var request createRequest
	if err := binding.Bind(ctx, &request); err != nil {
		status := http.StatusBadRequest
		message := "The form could not be read. Try again."
		if errors.Is(err, binding.ErrUnsupportedMediaType) {
			status = http.StatusUnsupportedMediaType
			message = err.Error()
		}
		form := newForm(request.Name, setup)
		form.SummaryError = message
		return h.formFailure(ctx, status, form, message)
	}
	create := h.service.Create
	if setup {
		create = h.service.Setup
	}
	profile, err := create(
		ctx.Request().Context(),
		&models.CreateParams{Name: request.Name},
	)
	if err != nil {
		return h.creationFailure(ctx, newForm(request.Name, setup), err)
	}
	if hypermedia.Negotiate(ctx.Request()) == hypermedia.RepresentationHTML {
		location := "/profiles"
		if setup {
			location = "/chores"
		}
		return hypermedia.HTMLRedirect(ctx, http.StatusSeeOther, location)
	}
	return hypermedia.JSON(ctx, http.StatusCreated, newResponse(profile))
}

func newForm(name string, setup bool) viewmodels.Form {
	form := viewmodels.Form{
		ActionHref: "/profiles", CancelHref: "/profiles",
		Name: viewmodels.Field{Value: name}, Setup: setup,
	}
	if setup {
		form.ActionHref = "/setup"
		form.CancelHref = ""
	}
	return form
}
